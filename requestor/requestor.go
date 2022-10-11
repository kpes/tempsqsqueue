package requestor

import (
	"context"
	"encoding/json"
	"tempqueue/common"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
)

type Requestor struct {
	client           *sqs.Client
	responseQueueUrl string
	waitTime         int32
}

type RequestorOption func(*Requestor)

func WithWaitTime(waitTime int32) RequestorOption {
	return func(h *Requestor) {
		h.waitTime = waitTime
	}
}

func NewRequestor(sqsClient *sqs.Client, responseQueueUrl string, opts ...RequestorOption) *Requestor {
	r := &Requestor{
		client:           sqsClient,
		responseQueueUrl: responseQueueUrl,
		waitTime:         10,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (t *Requestor) SendMessageAndWaitForResponse(ctx context.Context, sendToQueueUrl string, message any) (string, error) {
	correlationId := uuid.NewString()

	body, err := json.Marshal(message)
	if err != nil {
		return "", err
	}

	inputMessage := sqs.SendMessageInput{
		QueueUrl:    aws.String(sendToQueueUrl),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			common.CorrelationMessageAttributeKey: {
				StringValue: aws.String(correlationId),
			},
			common.ResponseQueueUrl: {
				StringValue: aws.String(t.responseQueueUrl),
			},
		},
	}

	_, err = t.client.SendMessage(ctx, &inputMessage)
	if err != nil {
		return "", err
	}

	result, err := t.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:        aws.String(t.responseQueueUrl),
		WaitTimeSeconds: t.waitTime,
	})
	if err != nil {
		return "", err
	}

	for _, message := range result.Messages {
		if val, ok := message.MessageAttributes[common.CorrelationMessageAttributeKey]; ok {
			if *val.StringValue == correlationId {
				return *message.Body, nil
			}
		}
	}

	return "", nil
}
