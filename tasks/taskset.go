package tasks

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

// TaskSet holds a chain of task to be carried out and a process configuration for
// how they should be carried out.
type TaskSet struct {
	NameID        string        `json:"name_id"        dynamodbav:"name_id"`
	Tasks         []Task        `json:"tasks"          dynamodbav:"tasks"`
	ProcessConfig ProcessConfig `json:"process_config" dynamodbav:"process_config"`
}

// Task is a single unit of work that can be batched together for form a chain of tasks.
type Task struct {
	Action string `json:"action" dynamodbav:"action"`
}

// ProcessConfig holds the configurations for how the job process should be set up.
type ProcessConfig struct {
	LambdaConcurrencyLimit   int `json:"concurrency_limit"           dynamodbav:"concurrency_limit"`
	LambdaTimeoutSes         int `json:"lambda_timeout_secs"         dynamodbav:"lambda_timeout_secs"`
	SQSVisibilityTimeoutSecs int `json:"sqs_visibility_timeout_secs" dynamodbav:"sqs_visibility_timeout_secs"`
}

// ListTaskSets Lists all the TaskSets in the table.
// NOTE: do not use Scan operation in production! This is a very expensive call,
// that is being used only for demo purposes for simplicity.
func ListTaskSets(db *dynamodb.DynamoDB, tableName string) ([]TaskSet, error) {
	out, err := db.Scan(&dynamodb.ScanInput{TableName: aws.String(tableName)})
	if err != nil {
		return nil, errors.Wrapf(err, "could not list task sets for table %s", tableName)
	}
	var taskSets []TaskSet
	if err := dynamodbattribute.UnmarshalListOfMaps(out.Items, &taskSets); err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal task sets from table %s", tableName)
	}
	return taskSets, nil
}
