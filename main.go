package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, event interface{}) (string, error) {
	fmt.Println("event", event)
	return "Hello world", nil
}

func main() {
	lambda.Start(handleRequest)
}