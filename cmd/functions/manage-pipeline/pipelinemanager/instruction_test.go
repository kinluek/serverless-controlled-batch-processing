package pipelinemanager_test

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/kinluek/serverless-controlled-batch-processing/cmd/functions/manage-pipeline/pipelinemanager"
	"github.com/kinluek/serverless-controlled-batch-processing/pipelineconfig"
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
				Config: &pipelineconfig.PipelineConfig{
					ID:                       "new-config-id",
					LambdaConcurrencyLimit:   5,
					LambdaTimeoutSes:         10,
					SQSVisibilityTimeoutSecs: 15,
				},
			},
		},
		{
			name: "update",
			file: "../../../../testdata/lambda-events/dynamodb-event-update-config.json",
			want: pipelinemanager.Instruction{
				Operation: pipelinemanager.Update,
				Config: &pipelineconfig.PipelineConfig{
					ID:                       "update-config-id",
					LambdaConcurrencyLimit:   12,
					LambdaTimeoutSes:         5,
					SQSVisibilityTimeoutSecs: 0,
				},
			},
		},
		{
			name: "delete",
			file: "../../../../testdata/lambda-events/dynamodb-event-delete-config.json",
			want: pipelinemanager.Instruction{
				Operation: pipelinemanager.Delete,
				Config: &pipelineconfig.PipelineConfig{
					ID: "delete-config-id",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := getRecordFromFile(t, tt.file)
			instruction, err := pipelinemanager.MakeInstructionFromStreamRecord(record)
			if err != nil {
				t.Fatalf("could not parse instruction from record: %v", err)
			}
			assert.Equal(t, instruction, tt.want)
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
