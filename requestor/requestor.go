package requestor

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/kpes/tempsqsqueue/internal/common"
)

type Requestor struct {
	client           sqsClient
	responseQueueUrl string
	waitTime         int32
}

type sqsClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
}

type RequestorOption func(*Requestor)

func WithWaitTime(waitTime int32) RequestorOption {
	return func(h *Requestor) {
		h.waitTime = waitTime
	}
}

func NewRequestor(client sqsClient, responseQueueUrl string, opts ...RequestorOption) *Requestor {
	r := &Requestor{
		client:           client,
		responseQueueUrl: responseQueueUrl,
		waitTime:         10,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

func (r *Requestor) SendMessageAndWaitForResponse(ctx context.Context, sendToQueueUrl string, message any, response any) error {
	correlationId := uuid.NewString()

	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	inputMessage := sqs.SendMessageInput{
		QueueUrl:    aws.String(sendToQueueUrl),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			common.CorrelationMessageAttributeKey: {
				StringValue: aws.String(correlationId),
			},
			common.ResponseQueueUrl: {
				StringValue: aws.String(r.responseQueueUrl),
			},
		},
	}

	_, err = r.client.SendMessage(ctx, &inputMessage)
	if err != nil {
		return err
	}

	received := false
	endTime := time.Now().Add(time.Second * time.Duration(r.waitTime))
	for !received && time.Now().Before(endTime) {
		result, err := r.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:        aws.String(r.responseQueueUrl),
			WaitTimeSeconds: 5,
		})
		if err != nil {
			return err
		}

		for _, message := range result.Messages {
			if val, ok := message.MessageAttributes[common.CorrelationMessageAttributeKey]; ok {
				if *val.StringValue == correlationId {
					if err = json.Unmarshal([]byte(*message.Body), &response); err != nil {
						return err
					}

					return nil
				}
			}
		}
	}

	return nil
}
