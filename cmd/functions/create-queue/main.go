package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/cmd/functions/create-queue/cqhandler"
	"github.com/kinluek/serverless-controlled-batch-processing/env"
	"github.com/kinluek/serverless-controlled-batch-processing/processconfigs"
	"github.com/pkg/errors"
)

// config holds the configuration parameters for the application.
type config struct {
	processConfigsTableName string // env:DYNAMO_PROCESS_CONFIGS_TABLE_NAME
	envName                 string // env:ENV_NAME
}

var cfg config
var sess *session.Session
var dbSvc *dynamodb.DynamoDB
var sqsSvc *sqs.SQS

// use init function to save on reinitialisation costs on lambda warm starts.
func init() {
	sess = session.Must(session.NewSession())
	dbSvc = dynamodb.New(sess)
	sqsSvc = sqs.New(sess)
	cfg = loadConfig()
}

// The Lambda function to be triggered when a new ProcessConfig is added to the job config table. It handles the setup
// of the new job queue along with a DLQ for the ProcessConfig.
func handle(ctx context.Context, event events.DynamoDBEvent) error {
	jobConfig, err := parseJobConfig(event)
	if err != nil {
		return errors.Wrapf(err, "could not parse dynamo event")
	}
	if err := cqhandler.New(dbSvc, sqsSvc, cfg.processConfigsTableName, cfg.envName).Handle(ctx, jobConfig); err != nil {
		return errors.Wrapf(err, "failed to queue creation for job config %s", jobConfig.ID)
	}
	return nil
}

func main() {
	lambda.Start(handle)
}

// loadConfig loads the environment variables needed for the configuration
// of the program, any problems with loading the config will cause the program to panic.
func loadConfig() config {
	const (
		EnvarTableName = "DYNAMO_PROCESS_CONFIGS_TABLE_NAME"
		EnvarEnvName   = "ENV_NAME"
	)
	tableName, err := env.GetEnvRequired(EnvarTableName)
	if err != nil {
		panic(err)
	}
	envName, err := env.GetEnvRequired(EnvarEnvName)
	if err != nil {
		panic(err)
	}
	return config{tableName, envName}
}

func parseJobConfig(event events.DynamoDBEvent) (processconfigs.ProcessConfig, error) {
	return processconfigs.ParseNewRecord(event.Records[0])
}
