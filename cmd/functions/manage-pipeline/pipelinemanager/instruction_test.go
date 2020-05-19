package pipelinemanager_test

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/kinluek/serverless-controlled-batch-processing/cmd/functions/manage-pipeline/pipelinemanager"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMakeInstructionFromStreamRecord(t *testing.T) {
	tests := []struct {
		name string
		file string
		want pipelinemanager.Instruction
	}{
		{
			name: "new",
			file: "../../../../testdata/lambda-events/dynamodb-event-new-config.json",
			want: pipelinemanager.Instruction{
				Operation: pipelinemanager.Add,
				Config: pipelinemanager.ConfigParams{
					ID:                       "new-config-id",
					LambdaConcurrencyLimit:   pInt(5),
					LambdaTimeoutSecs:        pInt(10),
					SQSVisibilityTimeoutSecs: pInt(15),
				},
			},
		},
		{
			name: "update",
			file: "../../../../testdata/lambda-events/dynamodb-event-update-config.json",
			want: pipelinemanager.Instruction{
				Operation: pipelinemanager.Update,
				Config: pipelinemanager.ConfigParams{
					ID:                     "update-config-id",
					LambdaConcurrencyLimit: pInt(12),
					LambdaTimeoutSecs:      pInt(5),
				},
			},
		},
		{
			name: "delete",
			file: "../../../../testdata/lambda-events/dynamodb-event-delete-config.json",
			want: pipelinemanager.Instruction{
				Operation: pipelinemanager.Delete,
				Config: pipelinemanager.ConfigParams{
					ID: "delete-config-id",
				},
			},
		},
		{
			name: "update with missing field",
			file: "../../../../testdata/lambda-events/dynamodb-event-update-missing-field.json",
			want: pipelinemanager.Instruction{
				Operation: pipelinemanager.Update,
				Config: pipelinemanager.ConfigParams{
					ID:                "update-config-id",
					LambdaTimeoutSecs: pInt(5),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := getRecordFromFile(t, tt.file)
			instruction, err := pipelinemanager.MakeInstruction(record, pipelinemanager.Constants{})
			if err != nil {
				t.Fatalf("could not parse instruction from record: %v", err)
			}
			assert.Equal(t, tt.want, instruction)
		})
	}
}

func getRecordFromFile(t *testing.T, filePath string) events.DynamoDBEventRecord {
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("could not open file %v: %v", filePath, err)
	}
	var e events.DynamoDBEvent
	d := json.NewDecoder(f)
	if err := d.Decode(&e); err != nil {
		t.Fatalf("failed to decode file %v: %v", filePath, err)
	}
	return e.Records[0]
}

func pInt(i int) *int {
	return &i
}
