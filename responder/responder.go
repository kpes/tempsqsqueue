package responder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/kpes/tempsqsqueue/internal/common"
)

type ProcessFunc func(message string) (string, error)

type Responder struct {
	client      sqsClient
	processFunc ProcessFunc
}

type sqsClient interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

func NewResponder(sqsClient sqsClient, processFunc ProcessFunc) Responder {
	return Responder{
		client:      sqsClient,
		processFunc: processFunc,
	}
}

func (r *Responder) ProcessAndReply(ctx context.Context, message types.Message) error {
	responseUrl, ok := message.MessageAttributes[common.ResponseQueueUrl]
	if !ok || len(*responseUrl.StringValue) == 0 {
		return &RequiredAttributeMissing{field: common.ResponseQueueUrl}
	}

	correlationId, ok := message.MessageAttributes[common.CorrelationMessageAttributeKey]
	if !ok || len(*correlationId.StringValue) == 0 {
		return &RequiredAttributeMissing{field: common.CorrelationMessageAttributeKey}
	}

	result, err := r.processFunc(*message.Body)
	if err != nil {
		return err
	}

	_, err = r.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    responseUrl.StringValue,
		MessageBody: aws.String(result),
		MessageAttributes: map[string]types.MessageAttributeValue{
			common.CorrelationMessageAttributeKey: correlationId,
		},
	})

	return err
}
