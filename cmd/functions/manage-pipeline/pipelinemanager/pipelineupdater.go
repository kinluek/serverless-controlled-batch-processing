package pipelinemanager

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/pipeline"
	"github.com/kinluek/serverless-controlled-batch-processing/pipeline/consumer"
	"github.com/kinluek/serverless-controlled-batch-processing/pipeline/queue"
	"github.com/pkg/errors"
)

// pipelineUpdater is responsible for updating pipelines
type pipelineUpdater struct {
	lambdaSvc *lambda.Lambda
	sqsSvc    *sqs.SQS
	db        *dynamodb.DynamoDB
}

func newUpdater(lamSvc *lambda.Lambda, sqsSvc *sqs.SQS, db *dynamodb.DynamoDB) *pipelineUpdater {
	return &pipelineUpdater{
		lambdaSvc: lamSvc,
		sqsSvc:    sqsSvc,
		db:        db,
	}
}

func (u *pipelineUpdater) update(ctx context.Context, config ConfigParams, constants Constants) error {
	ident, err := u.getIdentifiers(ctx, config, constants)
	if err != nil {
		return errors.Wrapf(err, "failed to get identifiers for pipeline %s", config.ID)
	}
	if err := u.updateConsumer(ctx, config, ident); err != nil {
		return errors.Wrapf(err, "failed to update consumer for pipeline %s", config.ID)
	}
	if err := u.updateQueue(ctx, config, ident); err != nil {
		return errors.Wrapf(err, "failed to update queue for pipeline %s", config.ID)
	}
	return nil
}

func (u *pipelineUpdater) updateConsumer(ctx context.Context, config ConfigParams, ident pipeline.Identifier) error {
	return consumer.Update(ctx, u.lambdaSvc, consumer.UpdateParams{
		Name:        ident.ConsumerName,
		Concurrency: pInt64(config.LambdaConcurrencyLimit),
		Timeout:     pInt64(config.LambdaTimeoutSecs),
	})
}

func (u *pipelineUpdater) updateQueue(ctx context.Context, config ConfigParams, ident pipeline.Identifier) error {
	if config.SQSVisibilityTimeoutSecs == nil {
		return nil
	}
	return queue.UpdateVisibilityTimeout(ctx, u.sqsSvc, ident.QueueURL, *config.SQSVisibilityTimeoutSecs)
}

func (u *pipelineUpdater) getIdentifiers(ctx context.Context, config ConfigParams, constants Constants) (pipeline.Identifier, error) {
	return pipeline.GetIdentifier(ctx, u.db, constants.IdentifiersTable, config.ID)
}

func pInt64(i *int) *int64 {
	if i == nil {
		return nil
	}
	i64 := int64(*i)
	return  &i64
}
