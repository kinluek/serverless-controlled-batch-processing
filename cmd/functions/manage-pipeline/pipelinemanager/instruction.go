package pipelinemanager

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/kinluek/serverless-controlled-batch-processing/eventutil"
	"github.com/kinluek/serverless-controlled-batch-processing/pipelineconfig"
	"github.com/pkg/errors"
)

type operation string

const (
	// Available operations.
	Add    operation = "add"
	Update operation = "update"
	Delete operation = "delete"
)

// Instruction tells the PipelineManager how to mange the pipeline.
type Instruction struct {
	Operation operation
	Config    pipelineconfig.PipelineConfig
}

// MakeInstructionFromStreamRecord takes a DynamoDBEventRecord and makes an Instruction from it,
// which can be passed to a PipelineManager.
func MakeInstructionFromStreamRecord(record events.DynamoDBEventRecord) (Instruction, error) {
	newImage := record.Change.NewImage
	oldImage := record.Change.OldImage
	op := getOperationFromImages(newImage, oldImage)
	switch op {
	case Add:
		return makeInstructionAdd(newImage)
	case Update:
		return makeInstructionUpdate(newImage, oldImage)
	case Delete:
		return makeInstructionDelete(oldImage)
	default:
		return makeInstructionError()
	}
}

func makeInstructionAdd(newImage map[string]events.DynamoDBAttributeValue) (Instruction, error) {
	var config pipelineconfig.PipelineConfig
	if err := eventutil.UnmarshalDynamoAttrMap(newImage, &config); err != nil {
		return Instruction{}, err
	}
	return Instruction{Operation: Add, Config: config}, nil
}

func makeInstructionUpdate(newImage, oldImage map[string]events.DynamoDBAttributeValue) (Instruction, error) {
	var nc pipelineconfig.PipelineConfig
	if err := eventutil.UnmarshalDynamoAttrMap(newImage, &nc); err != nil {
		return Instruction{}, err
	}
	var oc pipelineconfig.PipelineConfig
	if err := eventutil.UnmarshalDynamoAttrMap(oldImage, &oc); err != nil {
		return Instruction{}, err
	}
	var uc pipelineconfig.PipelineConfig
	uc.ID = nc.ID
	if nc.LambdaConcurrencyLimit != oc.LambdaConcurrencyLimit {
		uc.LambdaConcurrencyLimit = nc.LambdaConcurrencyLimit
	}
	if nc.LambdaTimeoutSes != oc.LambdaTimeoutSes {
		uc.LambdaTimeoutSes = nc.LambdaTimeoutSes
	}
	if nc.SQSVisibilityTimeoutSecs != oc.SQSVisibilityTimeoutSecs {
		uc.SQSVisibilityTimeoutSecs = nc.SQSVisibilityTimeoutSecs
	}
	return Instruction{Operation: Update, Config: uc}, nil
}

func makeInstructionDelete(oldImage map[string]events.DynamoDBAttributeValue) (Instruction, error) {
	var oc pipelineconfig.PipelineConfig
	if err := eventutil.UnmarshalDynamoAttrMap(oldImage, &oc); err != nil {
		return Instruction{}, err
	}
	dc := pipelineconfig.PipelineConfig{ID: oc.ID}
	return Instruction{Operation: Delete, Config: dc}, nil
}

func makeInstructionError() (Instruction, error) {
	return Instruction{}, errors.New("unknown instruction operation")
}

func getOperationFromImages(newImage, oldImage map[string]events.DynamoDBAttributeValue) operation {
	if newImage != nil && oldImage == nil {
		return Add
	}
	if newImage != nil && oldImage != nil {
		return Update
	}
	return Delete
}
