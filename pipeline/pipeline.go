// Package pipeline holds the pipeline configuration and identifier types
// Config and Identifier pairs should have the same ID.
package pipeline

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

// Config holds the configurations for how the task processing pipeline should be set up.
// For simplicity, we will limit the configurable parameters to just these values. There are many more
// Parameters that could be added to the configuration.
type Config struct {
	ID                       string `json:"id"                          dynamodbav:"id"`
	LambdaConcurrencyLimit   int    `json:"concurrency_limit"           dynamodbav:"concurrency_limit"`
	LambdaTimeoutSes         int    `json:"lambda_timeout_secs"         dynamodbav:"lambda_timeout_secs"`
	SQSVisibilityTimeoutSecs int    `json:"sqs_visibility_timeout_secs" dynamodbav:"sqs_visibility_timeout_secs"`
}

// Identifier holds the resource identifiers for the pipeline.
type Identifier struct {
	ID                 string `json:"id"                    dynamodbav:"id"`
	QueueURL           string `json:"queue_url"             dynamodbav:"queue_url"`
	QueueARN           string `json:"queue_arn"             dynamodbav:"queue_arn"`
	DeadLetterQueueURL string `json:"dead_letter_queue_url" dynamodbav:"dead_letter_queue_url"`
	DeadLetterQueueARN string `json:"dead_letter_queue_arn" dynamodbav:"dead_letter_queue_arn"`
	ConsumerName       string `json:"consumer_name"         dynamodbav:"consumer_name"`
	ConsumerARN        string `json:"consumer_arn"          dynamodbav:"consumer_arn"`
}

// PutConfig puts a Config into the Dynamo DB table.
func PutConfig(ctx context.Context, db *dynamodb.DynamoDB, tableName string, config Config) error {
	c, err := dynamodbattribute.MarshalMap(config)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal config %s", config.ID)
	}
	_, err = db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item:      c,
		TableName: aws.String(tableName),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to config item into dynamo table %s", tableName)
	}
	return nil
}

// Put puts an Identifier into the DynamoDB table.
func PutIdentifier(ctx context.Context, db *dynamodb.DynamoDB, tableName string, ident Identifier) error {
	c, err := dynamodbattribute.MarshalMap(ident)
	if err != nil {
		return errors.Wrapf(err, "failed to marshal config %s", ident.ID)
	}
	_, err = db.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item:      c,
		TableName: aws.String(tableName),
	})
	if err != nil {
		return errors.Wrapf(err, "failed to put identifier %s into dynamo table %s", ident.ID, tableName)
	}
	return nil
}
