package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
)

type Request struct {
	Message string `json:"message"`
}

type Response struct {
	Message string `json:"message"`
}

func HandleRequest(ctx context.Context, request Request) (Response, error) {
	return Response{
		Message: "Processed: " + request.Message,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
