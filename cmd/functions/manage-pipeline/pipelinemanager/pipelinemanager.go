package pipelinemanager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/consumer"
	"github.com/kinluek/serverless-controlled-batch-processing/queue"
	"github.com/pkg/errors"
)

// HandlerFunc is a function that can handle a PipelineConfig.
type HandlerFunc func(ctx context.Context, instruction Instruction) error

// PipelineManager handles the creation, configuration, updating and removal of pipelines.
type PipelineManager struct {
	sqsSvc      *sqs.SQS
	lambdaSvc   *lambda.Lambda
	queueSuffix string
	mids        []Middleware
}

// New returns a new instance of PipelineManager.
func New(sqsSvc *sqs.SQS, lambdaSvc *lambda.Lambda, envName string) *PipelineManager {
	return &PipelineManager{
		sqsSvc:      sqsSvc,
		lambdaSvc:   lambdaSvc,
		queueSuffix: fmt.Sprintf("-%s-queue", envName),
	}
}

// Use attaches middleware to the Handler
func (h *PipelineManager) Use(mids ...Middleware) {
	h.mids = append(h.mids, mids...)
}

// Handle takes PipelineConfig and manages the operations accordingly.
func (h *PipelineManager) Handle(ctx context.Context, instruction Instruction) error {
	hf := wrapMiddleware(h.handle, h.mids...)
	return hf(ctx, instruction)
}

func (h *PipelineManager) handle(ctx context.Context, instruction Instruction) error {
	switch instruction.Operation {
	case Add:
		return h.addPipeline(ctx, instruction.Config, instruction.Constants)
	default:
		return nil
	}
}

func (h *PipelineManager) addPipeline(ctx context.Context, config ConfigParams, constants Constants) error {
	if err := validateAddConfig(config); err != nil {
		return errors.Wrapf(err, "failed to validate config %s", config.ID)
	}
	queueArn, err := h.addQueue(ctx, config)
	if err != nil {
		return errors.Wrapf(err, "failed to add queue for config %s", config.ID)
	}
	if err := h.addConsumer(ctx, config, constants, queueArn); err != nil {
		return errors.Wrapf(err, "failed to add consumer for config %s and queue arn %s", config.ID, queueArn)
	}
	return nil
}

func (h *PipelineManager) addQueue(ctx context.Context, config ConfigParams) (string, error) {
	return queue.CreateWithDLQ(ctx, h.sqsSvc, getQueueName(config.ID, h.queueSuffix), *config.SQSVisibilityTimeoutSecs)
}


func (h *PipelineManager) addConsumer(ctx context.Context, config ConfigParams, constants Constants, queueArn string) error {
	return  consumer.Add(ctx, h.lambdaSvc, consumer.AddParams{
		Bucket:      constants.ConsumerBucket,
		Key:         constants.ConsumerKey,
		Name:        config.ID,
		Concurrency: int64(*config.LambdaConcurrencyLimit),
		Timeout:     int64(*config.LambdaTimeoutSecs),
		RoleArn:     constants.ConsumerRole,
		QueueARN:    queueArn,
	})
}

func getQueueName(configID, suffix string) string {
	return fmt.Sprintf("%s%s", configID, suffix)
}
