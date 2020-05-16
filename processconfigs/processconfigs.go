package processconfigs

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

// ProcessConfig holds the configurations for how the task processing resources should be set up.
type ProcessConfig struct {
	ID                       string `json:"id"                          dynamodbav:"id"`
	LambdaConcurrencyLimit   int    `json:"concurrency_limit"           dynamodbav:"concurrency_limit"`
	LambdaTimeoutSes         int    `json:"lambda_timeout_secs"         dynamodbav:"lambda_timeout_secs"`
	SQSVisibilityTimeoutSecs int    `json:"sqs_visibility_timeout_secs" dynamodbav:"sqs_visibility_timeout_secs"`
}

// Put puts a new ProcessConfig into the Dynamo DB table.
func Put(ctx context.Context, db *dynamodb.DynamoDB, tableName string, config ProcessConfig) error {
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
func List(ctx context.Context, db *dynamodb.DynamoDB, tableName string) ([]ProcessConfig, error) {
	out, err := db.ScanWithContext(ctx, &dynamodb.ScanInput{TableName: aws.String(tableName)})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list configs in table %s", tableName)
	}
	var taskSets []ProcessConfig
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &taskSets); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal configs from table %s", tableName)
	}
	return taskSets, nil
}


// ParseNewRecord parses the new ProcessConfig from a DynamoDB event record.
func ParseNewRecord(record events.DynamoDBEventRecord) (ProcessConfig, error) {
	var config ProcessConfig
	if err := unmarshalStreamImage(record.Change.NewImage, &config); err != nil {
		return config, errors.Wrap(err, "failed to parse dynamo record")
	}
	return config, nil
}


func unmarshalStreamImage(attribute map[string]events.DynamoDBAttributeValue, out interface{}) error {
	attrMap := make(map[string]*dynamodb.AttributeValue)
	for k, v := range attribute {
		var attr dynamodb.AttributeValue
		bytes, err := v.MarshalJSON(); if err != nil {
			return err
		}
		if err := json.Unmarshal(bytes, &attr); err != nil {
			return err
		}
		attrMap[k] = &attr
	}
	return dynamodbattribute.UnmarshalMap(attrMap, out)
}
