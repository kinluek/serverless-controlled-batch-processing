package cqhandler

import (
	"context"
	"github.com/kinluek/serverless-controlled-batch-processing/processconfigs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	statusSuccess = "success"
	statusFail    = "fail"
)

type HandlerFunc func(ctx context.Context, config processconfigs.ProcessConfig) error

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

func Log(log *logrus.Logger) Middleware {

	// Middleware to return
	f := func(before HandlerFunc) HandlerFunc {

		// Wrapped handler
		h := func(ctx context.Context, config processconfigs.ProcessConfig) error {
			if err := before(ctx, config); err != nil {
				err := errors.Wrapf(err, "fail - handling queue setup for process config %s", config.ID)
				log.WithFields(getLogFields(config, statusFail)).Error(err.Error())
				return err
			}
			log.WithFields(getLogFields(config, statusSuccess)).Infof("success - handling queue setup for process config %s", config.ID)
			return nil
		}
		return h
	}
	return f
}



func getLogFields(config processconfigs.ProcessConfig, status string) logrus.Fields {
	return logrus.Fields{
		"process_config_id": config.ID,
		"status":            status,
	}
}
