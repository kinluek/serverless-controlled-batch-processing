package cqhandler

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/jobconfigs"
	"github.com/kinluek/serverless-controlled-batch-processing/queue"
)

// queueSuffix is appended to the end to the TaskSet ID to create the queue name.
const queueSuffix = "-queue"

// CreateQueueHandler handles queue creation for job configs.
type CreateQueueHandler struct {
	dbSvc               *dynamodb.DynamoDB
	sqsSvc              *sqs.SQS
	jobConfigsTableName string
}

// New returns a new instance of CreateQueueHandler.
func New(dbSvc *dynamodb.DynamoDB, sqsSvc *sqs.SQS, tableName string) *CreateQueueHandler {
	return &CreateQueueHandler{
		dbSvc:               dbSvc,
		sqsSvc:              sqsSvc,
		jobConfigsTableName: tableName,
	}
}

// Handle create a a queue from a job config.
func (h *CreateQueueHandler) Handle(ctx context.Context, config jobconfigs.JobConfig) error {
	return h.createQueue(ctx, config.ID)
}

func (h *CreateQueueHandler) createQueue(ctx context.Context, id string) error {
	return queue.CreateQueueWithDLQ(ctx, h.sqsSvc, getQueueName(id))
}

func getQueueName(taskSetID string) string {
	return fmt.Sprintf("%s%s", taskSetID, queueSuffix)
}
