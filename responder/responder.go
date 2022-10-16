package responder

import (
	"context"
	"tempqueue/common"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type ProcessFunc func(message string) (string, error)

type Responder struct {
	client      *sqs.Client
	processFunc ProcessFunc
}

func NewResponder(sqsClient *sqs.Client, processFunc ProcessFunc) Responder {
	return Responder{
		client:      sqsClient,
		processFunc: processFunc,
	}
}

func (r *Responder) ProcessAndReply(ctx context.Context, message types.Message) error {
	responseUrl := message.MessageAttributes[common.ResponseQueueUrl]
	correlationId := message.MessageAttributes[common.CorrelationMessageAttributeKey]

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
