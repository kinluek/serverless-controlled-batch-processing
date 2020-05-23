package eventutil

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// UnmarshalDynamoAttrMap knows how to unmarshal map[string]events.DynamoDBAttributeValue types.
func UnmarshalDynamoAttrMap(attribute map[string]events.DynamoDBAttributeValue, out interface{}) error {
	attrMap := make(map[string]*dynamodb.AttributeValue)
	for k, v := range attribute {
		var attr dynamodb.AttributeValue
		bytes, err := v.MarshalJSON()
		if err != nil {
			return err
		}
		if err := json.Unmarshal(bytes, &attr); err != nil {
			return err
		}
		attrMap[k] = &attr
	}
	return dynamodbattribute.UnmarshalMap(attrMap, out)
}
