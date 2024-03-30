# GO SWIM!

Go swim is a simple program that checks daily tide predictions in Vancouver before sending a notification with windows of time where the tide is above a comfortable height of 2.25M

![image](https://github.com/confy/go-swim/assets/4352706/8d69c0cb-026f-4880-b895-488014af8522)

## Why?
I go for a swim everyday - For a while I wouldn't check the tides and I would end up at a very shallow beach, ending up with a pretty terrible swim. For a while I would check the tides manually, but I thought it would be fun to automate the process.

## How?
The program is run via a scheduled Lambda function that runs every morning at 9am. It reaches out to the [Canadian Hydrographic Service](https://tides.gc.ca/en/web-services-offered-canadian-hydrographic-service) API to get the tide predictions for the day. It then filters the predictions to only show windows of time where the tide is above 2.25M. Finally, it sends a notification to my phone with [ntfy.sh](https://ntfy.sh/).

Deployment is done via terraform, and the code is written in Go. ie *GO* Swim!
