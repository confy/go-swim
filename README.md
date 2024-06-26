# GO SWIM!

Go swim is a simple program that checks tide predictions before sending alerts with the best time to swim.

![image](https://github.com/confy/go-swim/assets/4352706/45013af2-a2f9-4ab4-8ed5-d9ef22370b15)

## Why?
I go for a swim in the ocean everyday. For a while I wouldn't check the tides and would sometimes end up taking a pretty terrible swim at a very shallow beach. Lately I've been checking the tides manually, but I thought it would be fun to automate the process.

## How?
The program is run in a Lambda function, triggered every morning by an EventBridge scheduler. It reaches out to the [Canadian Hydrographic Service](https://tides.gc.ca/en/web-services-offered-canadian-hydrographic-service) API to get local tide predictions for the day. It then filters results into windows of time where the tide is above 2.25M. Finally, it sends a notification to my phone with [ntfy.sh](https://ntfy.sh/).

Deployment is done via terraform, and the code is written in Go. ie *GO* Swim!
