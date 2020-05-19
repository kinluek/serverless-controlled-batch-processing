package pipelinemanager

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/kinluek/serverless-controlled-batch-processing/eventutil"
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
	Config    ConfigParams
	Constants Constants
}

// ConfigParams represents pipeline configuration parameters, pointer fields are optional.
type ConfigParams struct {
	ID                       string `json:"id" required:"true"`
	LambdaConcurrencyLimit   *int   `json:"concurrency_limit,omitempty"`
	LambdaTimeoutSecs        *int   `json:"lambda_timeout_secs,omitempty"`
	SQSVisibilityTimeoutSecs *int   `json:"sqs_visibility_timeout_secs,omitempty"`
}

// Constants are the application constant parameters.
type Constants struct {
	ConsumerBucket string
	ConsumerKey    string
	ConsumerRole   string
	EnvName        string
}

// MakeInstruction takes a DynamoDBEventRecord and a Constants object and makes an Instruction from it
// which can be passed to a PipelineManager.
func MakeInstruction(record events.DynamoDBEventRecord, constants Constants) (Instruction, error) {
	newImage := record.Change.NewImage
	oldImage := record.Change.OldImage
	op := getOperationFromImages(newImage, oldImage)
	switch op {
	case Add:
		return makeInstructionAdd(newImage, constants)
	case Update:
		return makeInstructionUpdate(newImage, oldImage, constants)
	case Delete:
		return makeInstructionDelete(oldImage, constants)
	default:
		return makeInstructionError()
	}
}

func makeInstructionAdd(newImage map[string]events.DynamoDBAttributeValue, constants Constants) (Instruction, error) {
	var config ConfigParams
	if err := eventutil.UnmarshalDynamoAttrMap(newImage, &config); err != nil {
		return Instruction{}, err
	}
	return Instruction{Operation: Add, Config: config, Constants: constants}, nil
}

func makeInstructionUpdate(newImage, oldImage map[string]events.DynamoDBAttributeValue, constants Constants) (Instruction, error) {
	var nc ConfigParams
	if err := eventutil.UnmarshalDynamoAttrMap(newImage, &nc); err != nil {
		return Instruction{}, err
	}
	var oc ConfigParams
	if err := eventutil.UnmarshalDynamoAttrMap(oldImage, &oc); err != nil {
		return Instruction{}, err
	}
	var uc ConfigParams
	uc.ID = nc.ID
	uc.LambdaConcurrencyLimit = getUpdatedInt(nc.LambdaConcurrencyLimit, oc.LambdaConcurrencyLimit)
	uc.LambdaTimeoutSecs = getUpdatedInt(nc.LambdaTimeoutSecs, oc.LambdaTimeoutSecs)
	uc.SQSVisibilityTimeoutSecs = getUpdatedInt(nc.SQSVisibilityTimeoutSecs, oc.SQSVisibilityTimeoutSecs)
	return Instruction{Operation: Update, Config: uc, Constants: constants}, nil
}

func makeInstructionDelete(oldImage map[string]events.DynamoDBAttributeValue, constants Constants) (Instruction, error) {
	var oc ConfigParams
	if err := eventutil.UnmarshalDynamoAttrMap(oldImage, &oc); err != nil {
		return Instruction{}, err
	}
	dc := ConfigParams{ID: oc.ID}
	return Instruction{Operation: Delete, Config: dc, Constants: constants}, nil
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

func getUpdatedInt(newInt, oldInt *int) *int {
	if newInt == nil {
		return nil
	}
	if oldInt == nil {
		return newInt
	}
	if *newInt != *oldInt {
		return newInt
	}
	return nil
}
