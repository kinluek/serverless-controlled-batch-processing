package queueinit

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kinluek/serverless-controlled-batch-processing/queue"
	"github.com/kinluek/serverless-controlled-batch-processing/tasks"
)

// queueSuffix is appended to the end to the TaskSet ID to create the queue name.
const queueSuffix = "-queue"

// SetupTaskQueuesError holds the individual queue creation errors for each task set.
// Handles these errors accordingly.
type SetupTaskQueuesError struct {
	Errors []*CreateQueueError
}

func (e *SetupTaskQueuesError) Error() string {
	return "queue creation errors occurred while setting up queues"
}

// CreateQueueError is the error for an individual create creation failure for a task set.
type CreateQueueError struct {
	OriginalErr error
	Message     string
	TaskSetID   string
	QueueName   string
}

func (e *CreateQueueError) Error() string {
	return e.OriginalErr.Error()
}

// Handler handles initiating new task queues for the available task sets
// that need processing.
type Handler struct {
	dbSvc            *dynamodb.DynamoDB
	sqsSvc           *sqs.SQS
	taskSetTableName string
}

// New returns a new instance of Handler
func New(dbSvc *dynamodb.DynamoDB, sqsSvc *sqs.SQS, tableName string) *Handler {
	return &Handler{
		dbSvc:            dbSvc,
		sqsSvc:           sqsSvc,
		taskSetTableName: tableName,
	}
}

// Handle runs the queue initialising.
func (h *Handler) Handle(ctx context.Context) error {
	taskSets, err := tasks.ListTaskSets(h.dbSvc, h.taskSetTableName)
	if err != nil {
		return err
	}
	if err := h.setupTaskQueues(ctx, getTaskIDs(taskSets)); err != nil {
		return err
	}
	return nil
}

func (h *Handler) setupTaskQueues(ctx context.Context, taskSetIDs []string) error {
	var (
		numQueues       = len(taskSetIDs)
		errChan         = make(chan *CreateQueueError, numQueues)
		createQueueErrs []*CreateQueueError
	)
	for _, id := range taskSetIDs {
		go h.setupTaskQueue(ctx, id, errChan)
	}
	for i := 0; i < numQueues; i++ {
		if err := <-errChan; err != nil {
			createQueueErrs = append(createQueueErrs, err)
		}
	}
	if createQueueErrs != nil {
		return &SetupTaskQueuesError{createQueueErrs}
	}
	return nil
}

func (h *Handler) setupTaskQueue(ctx context.Context, id string, errChan chan<- *CreateQueueError) {
	queueName := createQueueName(id)
	if err := queue.CreateQueueWithDLQ(ctx, h.sqsSvc, queueName); err != nil {
		errChan <- &CreateQueueError{
			OriginalErr: err,
			Message:     err.Error(),
			TaskSetID:   id,
			QueueName:   queueName,
		}
		return
	}
	errChan <- nil
}

func createQueueName(taskSetID string) string {
	return fmt.Sprintf("%s%s", taskSetID, queueSuffix)
}

func getTaskIDs(taskSets []tasks.TaskSet) []string {
	ids := make([]string, len(taskSets))
	for i := range taskSets {
		ids[i] = taskSets[i].TaskID
	}
	return ids
}
