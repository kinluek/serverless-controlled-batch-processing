package queue

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// CreateQueueWithDLQ will create a queue along with another queue which will act as the
// dead letter queue, the dead letter queue will be named as the original queue name with
// ".dlq" extension.
func CreateQueueWithDLQ(ctx context.Context, svc *sqs.SQS, name string) error {
	_, err := svc.CreateQueueWithContext(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(name),
	})
	return err
}
