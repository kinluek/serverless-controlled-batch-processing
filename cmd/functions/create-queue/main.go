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
	"github.com/kinluek/serverless-controlled-batch-processing/jobconfigs"
	"github.com/pkg/errors"
)

// config holds the configuration parameters for the application.
type config struct {
	jobConfigsTableName string // env:DYNAMO_JOB_CONFIGS_TABLE_NAME
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

// The Lambda function to be triggered when a new JobConfig is added to the job config table. It handles the setup
// of the new job queue along with a DLQ for the JobConfig.
func handle(ctx context.Context, event events.DynamoDBEvent) error {
	jobConfig, err := parseJobConfig(event)
	if err != nil {
		return errors.Wrapf(err, "could not parse dynamo event")
	}
	if err := cqhandler.New(dbSvc, sqsSvc, cfg.jobConfigsTableName).Handle(ctx, jobConfig); err != nil {
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
	// Configuration Variable Names.
	const EnvarTableName = "DYNAMO_JOB_CONFIGS_TABLE_NAME"

	tableName, err := env.GetEnvRequired(EnvarTableName)
	if err != nil {
		panic(err)
	}
	return config{
		jobConfigsTableName: tableName,
	}
}

func parseJobConfig(event events.DynamoDBEvent) (jobconfigs.JobConfig, error) {
	return jobconfigs.ParseNewRecord(event.Records[0])
}
