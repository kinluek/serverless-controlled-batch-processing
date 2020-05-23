package pipelinemanager

import (
	"context"
	
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/kinluek/serverless-controlled-batch-processing/pipeline"
)

// pipelineRemover is removing pipelines
type pipelineRemover struct {
	lambdaSvc *lambda.Lambda
	sqsSvc    *sqs.SQS
	db        *dynamodb.DynamoDB
	envName   string
}

func newRemover(lamSvc *lambda.Lambda, sqsSvc *sqs.SQS, db *dynamodb.DynamoDB, envName string) *pipelineRemover {
	return &pipelineRemover{
		lambdaSvc: lamSvc,
		sqsSvc:    sqsSvc,
		db:        db,
		envName:   envName,
	}
}

func (a *pipelineRemover) remove(ctx context.Context, config ConfigParams, constants Constants) error {

}

func (a *pipelineRemover) getIdentifiers(ctx context.Context, constants Constants, id string) (pipeline.Identifier, error) {

}
