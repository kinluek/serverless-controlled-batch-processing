package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/cmd/functions/initiate-task-queues/queueinit"
	"github.com/kinluek/serverless-controlled-batch-processing/conf"
	"github.com/pkg/errors"
)

// config holds the configuration parameters for the application.
type config struct {
	taskSetTableName string // env:DYNAMO_TASK_TABLE_NAME
}

var cfg config
var sess *session.Session
var dbSvc *dynamodb.DynamoDB
var sqsSvc *sqs.SQS

// use init function to save reinitialisation costs on lambda warm starts.
func init() {
	sess = session.Must(session.NewSession())
	dbSvc = dynamodb.New(sess)
	sqsSvc = sqs.New(sess)
	cfg = loadConfig()
}

func handle(ctx context.Context) error {
	if err := queueinit.New(dbSvc, sqsSvc, cfg.taskSetTableName).Handle(ctx); err != nil {
		return errors.Wrapf(err, "failed to initialise queues")
	}
	return nil
}

func main() {
	lambda.Start(handle)
}

// loadConfig loads the environment variables needed for the configuration
// of the program, any problems with loading the config will cause the program to panic.
func loadConfig() config {
	// Configuration Variable Names.
	const EnvarTableName = "DYNAMO_TASK_TABLE_NAME"

	tableName, err := conf.GetEnvRequired(EnvarTableName)
	if err != nil {
		panic(err)
	}
	return config{
		taskSetTableName: tableName,
	}
}
