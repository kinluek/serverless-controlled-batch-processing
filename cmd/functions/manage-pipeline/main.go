package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	lambdaHandler "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/cmd/functions/manage-pipeline/pipelinemanager"
	"github.com/kinluek/serverless-controlled-batch-processing/env"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
)

// getConstants loads constants from the environment.
func getConstants() pipelinemanager.Constants {
	const (
		EnvarEnvName          = "ENV_NAME"
		EnvarConsumerRole     = "CONSUMER_ROLE"
		EnvarConsumerBucket   = "CONSUMER_BUCKET"
		EnvarConsumerKey      = "CONSUMER_KEY"
		EnvarIdentifiersTable = "IDENTIFIERS_TABLE"
	)
	return pipelinemanager.Constants{
		EnvName:          getEnv(EnvarEnvName),
		ConsumerRole:     getEnv(EnvarConsumerRole),
		ConsumerBucket:   getEnv(EnvarConsumerBucket),
		ConsumerKey:      getEnv(EnvarConsumerKey),
		IdentifiersTable: getEnv(EnvarIdentifiersTable),
	}
}

var (
	constants pipelinemanager.Constants
	sess      *session.Session
	sqsSvc    *sqs.SQS
	lambdaSvc *lambda.Lambda
	db        *dynamodb.DynamoDB
	logger    *logrus.Logger
)

// use init function to save on reinitialisation costs on lambda warm starts.
func init() {
	constants = getConstants()
	sess = session.Must(session.NewSession())
	sqsSvc = sqs.New(sess)
	lambdaSvc = lambda.New(sess)
	db = dynamodb.New(sess)
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)
}

// The Lambda function to be triggered when changes happen on the pipeline configuration DynamoDB table.
func handle(ctx context.Context, event events.DynamoDBEvent) error {
	instruction, err := pipelinemanager.MakeInstruction(event.Records[0], constants)
	if err != nil {
		return errors.Wrap(err, "failed to make instruction from event event")
	}
	h := pipelinemanager.New(sqsSvc, lambdaSvc, db, constants.EnvName)
	h.Use(pipelinemanager.CatchPanic(logger))
	h.Use(pipelinemanager.Log(logger))
	return h.Handle(ctx, instruction)
}

func main() {
	lambdaHandler.Start(handle)
}

func getEnv(name string) string {
	val, err := env.GetEnvRequired(name)
	if err != nil {
		panic(err)
	}
	return val
}
