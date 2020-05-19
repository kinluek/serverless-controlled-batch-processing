package consumer

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
)

const (
	// use defaults for simplicity of the demo project.
	defaultBatchSize = 1
	defaultRuntime   = "go1.x"
	defaultHandler   = "consume"
	defaultEnabled   = true
)

// AddParams are the required parameters needed to Add a consumer
type AddParams struct {
	Bucket      string // S3 bucket name
	Key         string // S3 source code key
	Name        string // function name
	Concurrency int64  // concurrency limit of the function
	Timeout     int64  // function timeout in seconds
	RoleArn     string // ARN of the Lambda execution role
	QueueARN    string // ARN of the queue to consume
}

// Add adds a new consumer to an existing queue.
func Add(ctx context.Context, svc *lambda.Lambda, p AddParams) error {
	if err := createFunction(ctx, svc, p.Bucket, p.Key, p.Name, p.RoleArn, p.Timeout); err != nil {
		return errors.Wrapf(err, "failed to create function %s", p.Name)
	}
	if err := setConcurrency(ctx, svc, p.Name, p.Concurrency); err != nil {
		return errors.Wrapf(err, "failed to set function %s concurrency", p.Name)
	}
	if err := attachQueue(ctx, svc, p.Name, p.QueueARN); err != nil {
		return errors.Wrapf(err, "failed to attach consumer function %s to queue %s", p.Name, p.QueueARN)
	}
	return nil
}

func createFunction(ctx context.Context, svc *lambda.Lambda, bucket, key, name, roleArn string, timeout int64) error {
	_, err := svc.CreateFunctionWithContext(ctx, &lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			S3Bucket: aws.String(bucket),
			S3Key:    aws.String(key),
		},
		FunctionName: aws.String(name),
		Handler:      aws.String(defaultHandler),
		Role:         aws.String(roleArn),
		Runtime:      aws.String(defaultRuntime),
		Timeout:      aws.Int64(timeout),
	})
	return err
}

func setConcurrency(ctx context.Context, svc *lambda.Lambda, funcName string, concurrency int64) error {
	_, err := svc.PutFunctionConcurrencyWithContext(ctx, &lambda.PutFunctionConcurrencyInput{
		FunctionName:                 aws.String(funcName),
		ReservedConcurrentExecutions: aws.Int64(concurrency),
	})
	return err
}

func attachQueue(ctx context.Context, svc *lambda.Lambda, funcName, queueArn string) error {
	_, err := svc.CreateEventSourceMappingWithContext(ctx, &lambda.CreateEventSourceMappingInput{
		BatchSize:      aws.Int64(defaultBatchSize),
		Enabled:        aws.Bool(defaultEnabled),
		EventSourceArn: aws.String(queueArn),
		FunctionName:   aws.String(funcName),
	})
	return err
}
