package pipelinemanager

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"runtime/debug"
)

const (
	statusSuccess = "success"
	statusFail    = "fail"
)

// Middleware is a function that wraps a HandlerFunc to enhance its capabilities.
type Middleware func(h HandlerFunc) HandlerFunc

// wrapMiddleware applies the middleware so that they will be executed in the order they are provided.
func wrapMiddleware(handler HandlerFunc, middleware ...Middleware) HandlerFunc {

	// start wrapping the messageHandler from the end of the outerMiddleware
	// slice, to the start, this will ensure the code is executed in the right
	// order when, the resulting messageHandler is executed.
	for i := len(middleware) - 1; i >= 0; i-- {
		mw := middleware[i]
		if mw != nil {
			handler = mw(handler)
		}
	}
	return handler
}

// Log is a takes a logger and returns a middleware function that provides logging capabilities.
func Log(log *logrus.Logger) Middleware {

	// Middleware to return.
	return func(before HandlerFunc) HandlerFunc {

		// Handler to return.
		return func(ctx context.Context, instruction Instruction) error {
			const (
				msgTemplateFail    = "fail - handling queue setup for process instruction %s"
				msgTemplateSuccess = "success - handling queue setup for process instruction %s"
			)
			if err := before(ctx, instruction); err != nil {
				err := errors.Wrapf(err, msgTemplateFail, instruction.Config.ID)
				log.WithFields(getLogFields(instruction, statusFail)).Error(err.Error())
				return err
			}
			log.WithFields(getLogFields(instruction, statusSuccess)).Infof(msgTemplateSuccess, instruction.Config.ID)
			return nil
		}
	}

}

// CatchPanic stops panics from bubbling up and returns them as an error.
func CatchPanic(log *logrus.Logger) Middleware {

	// Middleware to return.
	return func(before HandlerFunc) HandlerFunc {

		// Handler to return.
		return func(ctx context.Context, instruction Instruction) (err error) {
			defer func() {
				if r := recover(); r != nil {
					if er, ok := r.(error); ok {
						err = errors.Wrapf(er, "panic occurred")
					}
					err = fmt.Errorf("panic occurred: %v", r)
					log.WithFields(getLogFields(instruction, statusFail)).Errorf("%s: stacktrace: \n%s", err, debug.Stack())
				}
			}()
			return before(ctx, instruction)
		}
	}
}

func getLogFields(instruction Instruction, status string) logrus.Fields {
	return logrus.Fields{
		"config_id": instruction.Config.ID,
		"operation":         instruction.Operation,
		"instruction":       instruction,
		"status":            status,
	}
}
