package responder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Responder struct {
	client *sqs.Client
}

func NewResponder(sqsClient *sqs.Client) Responder {
	return Responder{
		client: sqsClient,
	}
}

func (r *Responder) ProcessAndReply(ctx context.Context, message string) error {

	return nil
}
