package cqhandler

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/processconfigs"
	"github.com/kinluek/serverless-controlled-batch-processing/queue"
)

// CreateQueueHandler handles queue creation for the given process configs.
type CreateQueueHandler struct {
	dbSvc               *dynamodb.DynamoDB
	sqsSvc              *sqs.SQS
	jobConfigsTableName string
	queueSuffix         string
	mids                []Middleware
}

// New returns a new instance of CreateQueueHandler.
func New(dbSvc *dynamodb.DynamoDB, sqsSvc *sqs.SQS, tableName, envName string) *CreateQueueHandler {
	return &CreateQueueHandler{
		dbSvc:               dbSvc,
		sqsSvc:              sqsSvc,
		jobConfigsTableName: tableName,
		queueSuffix:         fmt.Sprintf("-%s-queue", envName),
	}
}

// Use attaches middleware to the Handler
func (h *CreateQueueHandler) Use(mids ...Middleware) {
	h.mids = append(h.mids, mids...)
}

// Handle create a a queue from a ProcessConfig.
func (h *CreateQueueHandler) Handle(ctx context.Context, config processconfigs.ProcessConfig) error {
	hf := wrapMiddleware(h.handle, h.mids...)
	return hf(ctx, config)
}

func (h *CreateQueueHandler) handle(ctx context.Context, config processconfigs.ProcessConfig) error {
	return h.createQueue(ctx, config.ID)
}

func (h *CreateQueueHandler) createQueue(ctx context.Context, id string) error {
	return queue.CreateQueueWithDLQ(ctx, h.sqsSvc, getQueueName(id, h.queueSuffix))
}

func getQueueName(processID, suffix string) string {
	return fmt.Sprintf("%s%s", processID, suffix)
}
