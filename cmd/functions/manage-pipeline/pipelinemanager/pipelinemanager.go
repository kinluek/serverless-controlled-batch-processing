package pipelinemanager

import (
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// HandlerFunc is a function that can handle a PipelineConfig.
type HandlerFunc func(ctx context.Context, instruction Instruction) error

// PipelineManager handles the creation, configuration, updating and removal of pipelines.
type PipelineManager struct {
	sqsSvc    *sqs.SQS
	lambdaSvc *lambda.Lambda
	db        *dynamodb.DynamoDB
	envName   string
	mids      []Middleware
}

// New returns a new instance of PipelineManager.
func New(sqsSvc *sqs.SQS, lambdaSvc *lambda.Lambda, db *dynamodb.DynamoDB, envName string) *PipelineManager {
	return &PipelineManager{
		sqsSvc:    sqsSvc,
		lambdaSvc: lambdaSvc,
		db:        db,
		envName:   envName,
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
		adder := newAdder(h.lambdaSvc, h.sqsSvc, h.db, h.envName)
		return adder.add(ctx, instruction.Config, instruction.Constants)
	default:
		return nil
	}
}
