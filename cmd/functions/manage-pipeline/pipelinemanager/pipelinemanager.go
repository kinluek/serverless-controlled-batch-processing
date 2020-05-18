package pipelinemanager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/queue"
)

// HandlerFunc is a function that can handle a PipelineConfig.
type HandlerFunc func(ctx context.Context, instruction Instruction) error

// PipelineManager handles the creation, configuration, updating and removal of pipelines.
type PipelineManager struct {
	sqsSvc      *sqs.SQS
	lambdaSvc   *lambda.Lambda
	tableName   string
	queueSuffix string
	mids        []Middleware
}

// New returns a new instance of PipelineManager.
func New(sqsSvc *sqs.SQS, lambdaSvc *lambda.Lambda, tableName, envName string) *PipelineManager {
	return &PipelineManager{
		sqsSvc:      sqsSvc,
		lambdaSvc:   lambdaSvc,
		tableName:   tableName,
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
		return h.addQueue(ctx, instruction.Config)
	default:
		return nil
	}
}

func (h *PipelineManager) addQueue(ctx context.Context, config ConfigParams) error {
	if err := validateAddConfig(config); err != nil {
		return err
	}
	return queue.CreateWithDLQ(ctx, h.sqsSvc, getQueueName(config.ID, h.queueSuffix), *config.SQSVisibilityTimeoutSecs)
}

func getQueueName(configID, suffix string) string {
	return fmt.Sprintf("%s%s", configID, suffix)
}
