package queue

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	"strconv"
)

const (
	attrNameQueueArn          = "QueueArn"
	attrNameRedrivePolicy     = "RedrivePolicy"
	attrNameVisibilityTimeout = "VisibilityTimeout"

	extensionDQL = "-dlq"

	defaultRedriveCount = 2
)

// CreateWithDLQ will create a queue along with another queue which will act as the
// dead letter queue, the dead letter queue will be named as the original queue name with
// "-dlq" suffix. The visibility timeout must also be provided.
func CreateWithDLQ(ctx context.Context, svc *sqs.SQS, name string, timeout int) error {
	dlqOutput, err := createQueue(ctx, svc, name+extensionDQL, nil)
	if err != nil {
		return errors.Wrapf(err, "creating dlq for %v", name)
	}
	dlqArn, err := getAttribute(ctx, svc, *dlqOutput.QueueUrl, attrNameQueueArn)
	if err != nil {
		return errors.Wrapf(err, "failed to get arn for dlq %v", *dlqOutput.QueueUrl)
	}
	attributes, err := makeAttributes(timeout, defaultRedriveCount, dlqArn)
	if err != nil {
		return err
	}
	if _, err := createQueue(ctx, svc, name, attributes); err != nil {
		return errors.Wrapf(err, "creating queue %v", name)
	}
	return nil
}

func createQueue(ctx context.Context, svc *sqs.SQS, name string, attributes map[string]*string) (*sqs.CreateQueueOutput, error) {
	return svc.CreateQueueWithContext(ctx, &sqs.CreateQueueInput{
		QueueName:  aws.String(name),
		Attributes: attributes,
	})
}

func getAttribute(ctx context.Context, svc *sqs.SQS, queueURL, attribute string) (string, error) {
	out, err := svc.GetQueueAttributesWithContext(ctx, &sqs.GetQueueAttributesInput{
		AttributeNames: []*string{aws.String(attribute)},
		QueueUrl:       aws.String(queueURL),
	})
	if err != nil {
		return "", err
	}
	if attr, ok := out.Attributes[attribute]; ok {
		return *attr, nil
	}
	return "", errors.Errorf("attribute %v does not exist", attribute)
}

func makeAttributes(visTimeout, receiveCount int, dlqArn string) (map[string]*string, error) {
	rdp, err := makeRedrivePolicy(receiveCount, dlqArn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to make redrive policy with queue arn %v", dlqArn)
	}
	return map[string]*string{
		attrNameRedrivePolicy:     aws.String(rdp),
		attrNameVisibilityTimeout: aws.String(strconv.Itoa(visTimeout)),
	}, nil
}

func makeRedrivePolicy(receiveCount int, dlqArn string) (string, error) {
	type redrivePolicy struct {
		MaxReceiveCount     int    `json:"maxReceiveCount"`
		DeadLetterTargetArn string `json:"deadLetterTargetArn"`
	}
	rdp := redrivePolicy{
		MaxReceiveCount:     receiveCount,
		DeadLetterTargetArn: dlqArn,
	}
	buf, err := json.Marshal(rdp)
	return string(buf), err
}
