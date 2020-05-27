package pipelinemanager

import (
	"context"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// pipelineUpdater is responsible for updating pipelines
type pipelineUpdater struct {
	lambdaSvc *lambda.Lambda
	sqsSvc    *sqs.SQS
	envName   string
}

func newUpdater(lamSvc *lambda.Lambda, sqsSvc *sqs.SQS) *pipelineUpdater {
	return &pipelineUpdater{
		lambdaSvc: lamSvc,
		sqsSvc:    sqsSvc,
	}
}

func (a *pipelineUpdater) update(ctx context.Context, config ConfigParams, constants Constants) error {
	return nil
}
