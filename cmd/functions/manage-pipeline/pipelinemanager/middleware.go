package pipelinemanager

import (
	"context"
	"github.com/kinluek/serverless-controlled-batch-processing/pipelineconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	statusSuccess = "success"
	statusFail    = "fail"
)

// Middleware is a function that wraps a HandlerFunc to enhance its capabilities.
type Middleware func(h HandlerFunc) HandlerFunc

// wrapMiddleware creates a new message handler by wrapping the middleware around
// the given handler. The middlewares will be executed  in the order they are provided.
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
	m := func(before HandlerFunc) HandlerFunc {

		// Handler to return.
		h := func(ctx context.Context, config pipelineconfig.PipelineConfig) error {
			const (
				msgTemplateFail    = "fail - handling queue setup for process config %s"
				msgTemplateSuccess = "success - handling queue setup for process config %s"
			)
			if err := before(ctx, config); err != nil {
				err := errors.Wrapf(err, msgTemplateFail, config.ID)
				log.WithFields(getLogFields(config, statusFail)).Error(err.Error())
				return err
			}
			log.WithFields(getLogFields(config, statusSuccess)).Infof(msgTemplateSuccess, config.ID)
			return nil
		}
		return h
	}
	return m
}

func getLogFields(config pipelineconfig.PipelineConfig, status string) logrus.Fields {
	return logrus.Fields{
		"process_config_id": config.ID,
		"status":            status,
	}
}