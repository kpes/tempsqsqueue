package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/kpes/tempsqsqueue/responder"
)

type SQSEvent struct {
	Records []types.Message
}

var r responder.Responder

func handler(ctx context.Context, event SQSEvent) error {
	for _, record := range event.Records {
		if err := r.ProcessAndReply(ctx, record); err != nil {
			log.Printf("error occured processing record %s", *record.MessageId)
		}
	}

	return nil
}

func processFunc(message string) (string, error) {
	return message, nil
}

func main() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	sqsClient := sqs.NewFromConfig(cfg)

	r = responder.NewResponder(sqsClient, processFunc)

	lambda.Start(handler)
}
