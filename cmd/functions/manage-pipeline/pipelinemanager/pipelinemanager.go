package pipelinemanager

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/pipelineconfig"
	"github.com/kinluek/serverless-controlled-batch-processing/queue"
)

// HandlerFunc is a function that can handle a PipelineConfig.
type HandlerFunc func(ctx context.Context, config pipelineconfig.PipelineConfig) error

// PipelineManager handles the creation, configuration, updating and removal of pipelines.
type PipelineManager struct {
	dbSvc       *dynamodb.DynamoDB
	sqsSvc      *sqs.SQS
	tableName   string
	queueSuffix string
	mids        []Middleware
}

// New returns a new instance of PipelineManager.
func New(dbSvc *dynamodb.DynamoDB, sqsSvc *sqs.SQS, tableName, envName string) *PipelineManager {
	return &PipelineManager{
		dbSvc:       dbSvc,
		sqsSvc:      sqsSvc,
		tableName:   tableName,
		queueSuffix: fmt.Sprintf("-%s-queue", envName),
	}
}

// Use attaches middleware to the Handler
func (h *PipelineManager) Use(mids ...Middleware) {
	h.mids = append(h.mids, mids...)
}

// Handle takes PipelineConfig and manages the operations accordingly.
func (h *PipelineManager) Handle(ctx context.Context, config pipelineconfig.PipelineConfig) error {
	hf := wrapMiddleware(h.handle, h.mids...)
	return hf(ctx, config)
}

func (h *PipelineManager) handle(ctx context.Context, config pipelineconfig.PipelineConfig) error {
	return h.createQueue(ctx, config.ID)
}

func (h *PipelineManager) createQueue(ctx context.Context, id string) error {
	return queue.CreateWithDLQ(ctx, h.sqsSvc, getQueueName(id, h.queueSuffix))
}

func getQueueName(configID, suffix string) string {
	return fmt.Sprintf("%s%s", configID, suffix)
}
