package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"time"
)

// handle is the task process which consumers the queue.
// For now all it will do is sleep for a second and then print the SQS message.
func handle(event events.SQSEvent) error {
	time.Sleep(time.Second)
	buf, err := json.Marshal(event.Records[0])
	if err != nil {
		return errors.Wrap(err, "failed to marshal event")
	}
	fmt.Println(string(buf))
	return nil
}

func main() {
	lambda.Start(handle)
}
