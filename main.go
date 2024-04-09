package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	API_URL = "https://api.iwls-sine.azure.cloud-nuage.dfo-mpo.gc.ca/api/v1"
	STATION_ID = "5cebf1de3d0f4a073c4bb943"
	RESOLUTION = "FIVE_MINUTES"
	TZ = "America/Vancouver"
	MinumumTideHeightMeters = 2.25

)

type TideEvent struct {
	EventDate time.Time `json:"eventDate"`
	QcFlagCode string `json:"qcFlagCode"`
	Value float64 `json:"value"`
	TimeSeriesId string `json:"timeSeriesId"`
}

type HiLoTides struct {
	HighTides []TideEvent
	LowTides  []TideEvent
}

type SwimWindow struct {
	StartTime time.Time
	EndTime time.Time
}


type Output struct {
	Title string
	Content string
	Tags []string
}

func main() {
	lambda.Start(goSwim)
	// goSwim()
}

func goSwim() (Output, error) {
	tz, err := time.LoadLocation(TZ)
	if err != nil {
		fmt.Println("Error loading timezone: ", err)
		panic(err)
	}

	t := time.Now()
	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, tz).In(time.UTC)
	endTime := startTime.AddDate(0, 0, 1)

	tideEvents, err := getStationTidePredictions(STATION_ID, startTime, endTime)
	if err != nil {
		fmt.Println("Error getting tide predictions: ", err)
		panic(err)
	}
	swimWindows := createSwimWindowsFromTides(tideEvents)

	hiLoTideEvents, err := getStationHiLoTide(STATION_ID, startTime, endTime)
	if err != nil {
		fmt.Println("Error getting hi lo tides: ", err)
		panic(err)
	}
	hiLoTides := formatHiLoTides(hiLoTideEvents)

	output := formatOutput(startTime, hiLoTides, tz, swimWindows)

	req, _ := http.NewRequest("POST", "https://ntfy.sh/go-swim-vancouver", strings.NewReader(output.Content))
	req.Header.Set("Title", output.Title)
	req.Header.Set("Tags", strings.Join(output.Tags, ","))
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending notification: ", err)
		panic(err)
	}
	return output, nil
}

// Make a request to the station API and return TideEvent structs
func makeStationRequest(reqURL string) ([]TideEvent, error) {
	httpClient := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %v", err)
	}
	req.Header.Add("accept", "*/*")
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error making request: %v", err)
	}
	var tideEvents []TideEvent
	err = json.NewDecoder(resp.Body).Decode(&tideEvents)
	if err != nil {
		return nil, fmt.Errorf("Error decoding response: %v", err)
	}
	return tideEvents, nil
}

// Get the high and low tide events
func getStationHiLoTide(stationID string, startTime time.Time, endTime time.Time) ([]TideEvent, error) {
	reqUrl := fmt.Sprintf("%s/stations/%s/data?time-series-code=wlp-hilo&from=%s&to=%s", API_URL, stationID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
	tideEvents, err := makeStationRequest(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("Error getting station data: %v", err)
	}
	return tideEvents, nil
}

// Get the tide predictions for every interval
func getStationTidePredictions(stationID string, startTime time.Time, endTime time.Time) ([]TideEvent, error) {
	reqURL := fmt.Sprintf("%s/stations/%s/data?time-series-code=wlp&from=%s&to=%s&resolution=%s", API_URL, stationID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), RESOLUTION)
	tideEvents, err := makeStationRequest(reqURL)
	if err != nil {
		return nil, fmt.Errorf("Error getting station data: %v", err)
	}
	return tideEvents, nil
}

// Format the hi lo tide events into sorted high and low tides
func formatHiLoTides(hiLoTideEvents []TideEvent) HiLoTides {
	var hiLoTides HiLoTides
	var hiTideIndex int

	if hiLoTideEvents[0].Value > hiLoTideEvents[1].Value {
		// first event is a high tide
		hiTideIndex = 0	
	}	else {
		// first event is a low tide
		hiTideIndex = 1
	}

	for i := 0; i < len(hiLoTideEvents); i++ {
		if i % 2 == hiTideIndex {
			hiLoTides.HighTides = append(hiLoTides.HighTides, hiLoTideEvents[i])
		} else {
			hiLoTides.LowTides = append(hiLoTides.LowTides, hiLoTideEvents[i])
		}
	}

	return hiLoTides
}

// Create swim windows above the minimum tide height
func createSwimWindowsFromTides(tideEvents []TideEvent) []SwimWindow {
	var swimWindows []SwimWindow
	var swimWindow *SwimWindow

	for _, event := range tideEvents {
		if event.Value >= MinumumTideHeightMeters {
			if swimWindow == nil {
				swimWindow = &SwimWindow{StartTime: event.EventDate}
			}
		} else if swimWindow != nil {
			swimWindow.EndTime = event.EventDate
			swimWindows = append(swimWindows, *swimWindow)
			swimWindow = nil
		}
	}

	if swimWindow != nil {
		swimWindow.EndTime = tideEvents[len(tideEvents)-1].EventDate
		swimWindows = append(swimWindows, *swimWindow)
	}
	return swimWindows
}

// Format the output struct for ntfy.sh
func formatOutput(startTime time.Time, hiLoTides HiLoTides, tz *time.Location, swimWindows []SwimWindow) Output {
	dateOutput := startTime.Format("Mon January 2, 2006")
	metersOutput := strconv.FormatFloat(MinumumTideHeightMeters, 'f', 2, 64)

	output := Output{
		Title:   fmt.Sprintf("%s - GO SWIM!", dateOutput),
		Tags:    []string{"ocean", "swimmer"},
		Content: "",
	}

	output.Content += "High Tide: "
	for _, tide := range hiLoTides.HighTides {
		output.Content += fmt.Sprintf("%s ", tide.EventDate.In(tz).Format("15:04"))
	}

	output.Content += fmt.Sprintf("\nTide is higher than %s meters during:\n", metersOutput)
	for _, window := range swimWindows {
		output.Content += fmt.Sprintf("%s - %s\n", window.StartTime.In(tz).Format("15:04"), window.EndTime.In(tz).Format("15:04"))
	}
	// remove the last newline
	output.Content = output.Content[:len(output.Content)-1]
	return output
}
