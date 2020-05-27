package pipelinemanager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/pipeline"
	"github.com/kinluek/serverless-controlled-batch-processing/pipeline/consumer"
	"github.com/kinluek/serverless-controlled-batch-processing/pipeline/queue"
	"github.com/pkg/errors"
)

// pipelineAdder is responsible for adding pipelines
type pipelineAdder struct {
	lambdaSvc *lambda.Lambda
	sqsSvc    *sqs.SQS
	db        *dynamodb.DynamoDB
	envName   string
}

func newAdder(lamSvc *lambda.Lambda, sqsSvc *sqs.SQS, db *dynamodb.DynamoDB, envName string) *pipelineAdder {
	return &pipelineAdder{
		lambdaSvc: lamSvc,
		sqsSvc:    sqsSvc,
		db:        db,
		envName:   envName,
	}
}

func (a *pipelineAdder) add(ctx context.Context, config ConfigParams, constants Constants) error {
	if err := validateAddConfig(config); err != nil {
		return errors.Wrapf(err, "failed to validate config %s", config.ID)
	}
	queueOut, err := a.addQueue(ctx, config)
	if err != nil {
		return errors.Wrapf(err, "failed to add queue for config %s", config.ID)
	}
	consumerOut, err := a.addConsumer(ctx, config, constants, queueOut.Main.ARN)
	if err != nil {
		return errors.Wrapf(err, "failed to add consumer for config %s and queue arn %s", config.ID, queueOut.Main.ARN)
	}
	if err := a.addIdentifier(ctx, config, constants, queueOut, consumerOut); err != nil {
		return errors.Wrapf(err, "failed to add pipeline identifier for pipeline %s to table %s", config.ID, constants.IdentifiersTable)
	}
	return nil
}

func (a *pipelineAdder) addQueue(ctx context.Context, config ConfigParams) (queue.IdentifierPair, error) {
	return queue.CreateWithDLQ(ctx, a.sqsSvc, a.makeQueueName(config.ID), *config.SQSVisibilityTimeoutSecs)
}

func (a *pipelineAdder) addConsumer(ctx context.Context, config ConfigParams, constants Constants, queueArn string) (consumer.Identifier, error) {
	return consumer.Add(ctx, a.lambdaSvc, consumer.AddParams{
		Bucket:      constants.ConsumerBucket,
		Key:         constants.ConsumerKey,
		Name:        a.makeConsumerName(config.ID),
		Concurrency: int64(*config.LambdaConcurrencyLimit),
		Timeout:     int64(*config.LambdaTimeoutSecs),
		RoleArn:     constants.ConsumerRole,
		QueueARN:    queueArn,
	})
}

func (a *pipelineAdder) addIdentifier(ctx context.Context, config ConfigParams, constants Constants, qIdent queue.IdentifierPair, cIdent consumer.Identifier) error {
	ident := makePipelineIdentifier(config.ID, qIdent, cIdent)
	return pipeline.PutIdentifier(ctx, a.db, constants.IdentifiersTable, ident)
}

func (a *pipelineAdder) makeQueueName(id string) string {
	return fmt.Sprintf("%s-%s-queue", id, a.envName)
}

func (a *pipelineAdder) makeConsumerName(id string) string {
	return fmt.Sprintf("%s-%s-consumer", id, a.envName)
}

func validateAddConfig(config ConfigParams) error {
	if config.SQSVisibilityTimeoutSecs == nil {
		return errors.New("invalid add config: missing sqs visibility timeout")
	}
	if config.LambdaTimeoutSecs == nil {
		return errors.New("invalid add config: missing lambda timeout")
	}
	if config.LambdaConcurrencyLimit == nil {
		return errors.New("invalid add config: missing lambda concurrency")
	}
	return nil
}

func makePipelineIdentifier(id string, qi queue.IdentifierPair, ci consumer.Identifier) pipeline.Identifier {
	return pipeline.Identifier{
		ID:                 id,
		QueueURL:           qi.Main.URL,
		QueueARN:           qi.Main.ARN,
		DeadLetterQueueURL: qi.DLQ.URL,
		DeadLetterQueueARN: qi.DLQ.ARN,
		ConsumerName:       ci.Name,
		ConsumerARN:        ci.Arn,
	}
}
