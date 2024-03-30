package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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


type SwimWindow struct {
	StartTime time.Time
	EndTime time.Time
}


func handleRequest(ctx context.Context, event interface{}) (string, error) {
	PST, err := time.LoadLocation(TZ)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	t := time.Now()
	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, PST).In(time.UTC)
	endTime := startTime.AddDate(0, 0, 1)

	tideEvents, err := getStationTidePredictions(STATION_ID, startTime, endTime)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	swimWindows := createSwimWindowsFromTides(tideEvents)

	dateOutput := startTime.Format("Mon January 2, 2006")
	metersOutput := strconv.FormatFloat(MinumumTideHeightMeters, 'f', 2, 64)

	output := fmt.Sprintf("%s - GO SWIM!\nTides are higher than %s meters during:\n", dateOutput, metersOutput)

	for _, window := range swimWindows {
		window.StartTime = window.StartTime.In(PST)
		window.EndTime = window.EndTime.In(PST)
		windowString := fmt.Sprintf("%s - %s\n", window.StartTime.Format("15:04:05"), window.EndTime.Format("15:04:05"))
		output += windowString
	}

	fmt.Println(output)
	return output, nil
}


func getTimeWindow(timezone time.Location) (time.Time, time.Time) {
	t := time.Now()
	startTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, &timezone).In(time.UTC)
	endTime := startTime.AddDate(0, 0, 1)
	return startTime, endTime
}


func getStationTidePredictions(stationID string, startTime time.Time, endTime time.Time) ([]TideEvent, error) {
	reqURL := fmt.Sprintf("%s/stations/%s/data?time-series-code=wlp&from=%s&to=%s&resolution=%s", API_URL, stationID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339), RESOLUTION)
	httpClient := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	req.Header.Add("accept", "*/*")
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var tideEvents []TideEvent
	err = json.NewDecoder(resp.Body).Decode(&tideEvents)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return tideEvents, nil
}


func createSwimWindowsFromTides(tideEvents []TideEvent) []SwimWindow {
	var swimWindows []SwimWindow
	var swimWindow *SwimWindow

	for _, event := range tideEvents {
		if event.Value >= MinumumTideHeightMeters {
			if swimWindow == nil {
				swimWindow = &SwimWindow{StartTime: event.EventDate}
			}
		} else {
			if swimWindow != nil {
				swimWindow.EndTime = event.EventDate
				swimWindows = append(swimWindows, *swimWindow)
				swimWindow = nil
			}
		}
	}

	if swimWindow != nil {
		swimWindow.EndTime = tideEvents[len(tideEvents)-1].EventDate
		swimWindows = append(swimWindows, *swimWindow)
	}
	return swimWindows
}


func main() {
	lambda.Start(handleRequest)
}
