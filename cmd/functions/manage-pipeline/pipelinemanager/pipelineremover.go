package pipelinemanager

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/consumer"
	"github.com/kinluek/serverless-controlled-batch-processing/pipeline"
	"github.com/kinluek/serverless-controlled-batch-processing/queue"
	"github.com/pkg/errors"
)

// pipelineRemover is responsible for removing pipelines
type pipelineRemover struct {
	lambdaSvc *lambda.Lambda
	sqsSvc    *sqs.SQS
	db        *dynamodb.DynamoDB
}

func newRemover(lamSvc *lambda.Lambda, sqsSvc *sqs.SQS, db *dynamodb.DynamoDB) *pipelineRemover {
	return &pipelineRemover{
		lambdaSvc: lamSvc,
		sqsSvc:    sqsSvc,
		db:        db,
	}
}

func (r *pipelineRemover) remove(ctx context.Context, config ConfigParams, constants Constants) error {
	ident, err := r.getIdentifiers(ctx, config, constants)
	if err != nil {
		return errors.Wrapf(err, "failed to get identifiers for pipeline %s", config.ID)
	}
	if err := r.removeConsumer(ctx, ident); err != nil {
		return errors.Wrapf(err, "failed to remove consumer for pipeline %s", config.ID)
	}
	if err := r.removeQueue(ctx, ident); err != nil {
		return errors.Wrapf(err, "failed to remove queue for pipeline %s", config.ID)
	}
	if err := r.removeIdentifier(ctx, constants, ident); err != nil {
		return errors.Wrapf(err, "failed to remove identifier for pipeline %s", config.ID)
	}
}

func (r *pipelineRemover) getIdentifiers(ctx context.Context, config ConfigParams, constants Constants) (pipeline.Identifier, error) {
	return pipeline.GetIdentifier(ctx, r.db, constants.IdentifiersTable, config.ID)
}

func (r *pipelineRemover) removeConsumer(ctx context.Context, identifier pipeline.Identifier) error {
	return consumer.Delete(ctx, r.lambdaSvc, identifier.ConsumerName)
}

func (r *pipelineRemover) removeQueue(ctx context.Context, identifier pipeline.Identifier) error {
	if err := queue.Delete(ctx, r.sqsSvc, identifier.QueueURL); err != nil {
		return errors.Wrap(err, "failed to delete main queue")
	}
	if err := queue.Delete(ctx, r.sqsSvc, identifier.DeadLetterQueueURL); err != nil {
		return errors.Wrap(err, "failed to delete dead letter queue")
	}
	return nil
}

func (r *pipelineRemover) removeIdentifier(ctx context.Context, constants Constants, identifier pipeline.Identifier) error {
	return pipeline.DeleteItem(ctx, r.db, constants.IdentifiersTable, identifier.ID)
}
