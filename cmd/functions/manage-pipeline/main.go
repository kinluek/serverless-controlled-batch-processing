package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/cmd/functions/manage-pipeline/pipelinemanager"
	"github.com/kinluek/serverless-controlled-batch-processing/env"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

// envars holds the configuration parameters for the application.
type envars struct {
	tableName string // env:DYNAMO_PIPELINE_CONFIGS_TABLE_NAME
	envName   string // env:ENV_NAME
}

var (
	envs   envars
	sess   *session.Session
	dbSvc  *dynamodb.DynamoDB
	sqsSvc *sqs.SQS
	logger *logrus.Logger
)

// use init function to save on reinitialisation costs on lambda warm starts.
func init() {
	envs = loadEnvars()
	sess = session.Must(session.NewSession())
	dbSvc = dynamodb.New(sess)
	sqsSvc = sqs.New(sess)
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{PrettyPrint: true})
	logger.SetOutput(os.Stdout)
}

// The Lambda function to be triggered when changes happen on the pipeline configuration DynamoDB table.
func handle(ctx context.Context, event events.DynamoDBEvent) error {
	instruction, err := pipelinemanager.MakeInstructionFromStreamRecord(event.Records[0])
	if err != nil {
		return errors.Wrap(err, "failed to make instruction from event event")
	}
	h := pipelinemanager.New(dbSvc, sqsSvc, envs.tableName, envs.envName)
	h.Use(pipelinemanager.Log(logger))
	return h.Handle(ctx, instruction)
}

func main() {
	lambda.Start(handle)
}

// loadEnvars loads the environment variables needed for the configuration
// of the program, any problems with loading the envars will cause the program to panic.
func loadEnvars() envars {
	const (
		EnvarTableName = "DYNAMO_PIPELINE_CONFIGS_TABLE_NAME"
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
	return envars{tableName, envName}
}
