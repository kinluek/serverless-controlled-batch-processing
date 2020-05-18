package pipelinemanager

import "errors"

func validateAddConfig(config ConfigParams) error {
	if config.SQSVisibilityTimeoutSecs == nil {
		return errors.New("invalid add config: missing sqs visibility timeout")
	}
	if config.LambdaTimeoutSes == nil {
		return errors.New("invalid add config: missing lambda timeout")
	}
	if config.LambdaConcurrencyLimit == nil {
		return errors.New("invalid add config: missing lambda concurrency")
	}
	return nil
}
