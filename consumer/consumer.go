// Package consumer provides an API to add, update and delete consumers.
package consumer

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
	"time"
)

const (
	// use defaults for simplicity of the demo project.
	defaultBatchSize = 1
	defaultRuntime   = "go1.x"
	defaultHandler   = "consume"
	defaultEnabled   = true
	defaultWaitSecs  = 20
)

// Identifier holds the consumer identifiers, the lambda function name and ARN.
type Identifier struct {
	Name string
	Arn  string
}

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
func Add(ctx context.Context, svc *lambda.Lambda, p AddParams) (Identifier, error) {
	ident, err := createFunction(ctx, svc, p.Bucket, p.Key, p.Name, p.RoleArn, p.Timeout)
	if err != nil {
		return Identifier{}, errors.Wrapf(err, "failed to create function %s", p.Name)
	}
	if err := waitTillActive(ctx, svc, p.Name, defaultWaitSecs); err != nil {
		return Identifier{}, errors.Wrapf(err, "failed to wait for function %s to be active", p.Name)
	}
	if err := setConcurrency(ctx, svc, p.Name, p.Concurrency); err != nil {
		return Identifier{}, errors.Wrapf(err, "failed to set function %s concurrency", p.Name)
	}
	if err := attachQueue(ctx, svc, p.Name, p.QueueARN); err != nil {
		return Identifier{}, errors.Wrapf(err, "failed to attach consumer function %s to queue %s", p.Name, p.QueueARN)
	}
	return ident, nil
}

// Delete takes a function name and deletes it.
func Delete(ctx context.Context, svc *lambda.Lambda, name string) error {
	if _, err := svc.DeleteFunctionWithContext(ctx, &lambda.DeleteFunctionInput{FunctionName: aws.String(name)}); err != nil {
		return errors.Wrapf(err, "failed to delete function %s", name)
	}
	return nil
}

func createFunction(ctx context.Context, svc *lambda.Lambda, bucket, key, name, roleArn string, timeout int64) (Identifier, error) {
	output, err := svc.CreateFunctionWithContext(ctx, &lambda.CreateFunctionInput{
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
	if err != nil {
		return Identifier{}, err
	}
	return Identifier{*output.FunctionName, *output.FunctionArn}, nil
}

func waitTillActive(ctx context.Context, svc *lambda.Lambda, name string, waitSecs int) error {
	var state string
	for i := 0; i < waitSecs; i++ {
		c, err := svc.GetFunctionConfigurationWithContext(ctx, &lambda.GetFunctionConfigurationInput{
			FunctionName: aws.String(name),
		})
		if err != nil {
			return err
		}
		if *c.State == lambda.StateActive {
			return nil
		}
		state = *c.State
		time.Sleep(time.Second)
	}
	return errors.Errorf("function is %s after %v", state, waitSecs)
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
