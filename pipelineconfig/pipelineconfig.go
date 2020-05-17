package pipelineconfig

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

// PipelineConfig holds the configurations for how the task processing pipeline should be set up.
type PipelineConfig struct {
	ID                       string `json:"id"                          dynamodbav:"id"`
	LambdaConcurrencyLimit   int    `json:"concurrency_limit"           dynamodbav:"concurrency_limit"`
	LambdaTimeoutSes         int    `json:"lambda_timeout_secs"         dynamodbav:"lambda_timeout_secs"`
	SQSVisibilityTimeoutSecs int    `json:"sqs_visibility_timeout_secs" dynamodbav:"sqs_visibility_timeout_secs"`
}

// Put puts a new PipelineConfig into the Dynamo DB table.
func Put(ctx context.Context, db *dynamodb.DynamoDB, tableName string, config PipelineConfig) error {
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

// List Lists all the ProcessConfigs in the Dynamo DB table.
// NOTE: do not use Scan operation in production! This is a very expensive call,
// that is being used only for demo purposes for simplicity.
func List(ctx context.Context, db *dynamodb.DynamoDB, tableName string) ([]PipelineConfig, error) {
	out, err := db.ScanWithContext(ctx, &dynamodb.ScanInput{TableName: aws.String(tableName)})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list configs in table %s", tableName)
	}
	var taskSets []PipelineConfig
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &taskSets); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal configs from table %s", tableName)
	}
	return taskSets, nil
}
