package runner

import (
	"context"
	"github.com/sirupsen/logrus"
)

var loggerKey = "logger"

func NewLogger() logrus.FieldLogger {
	return logrus.StandardLogger()
}

func WithLogger(ctx context.Context, logger logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func WithLoggerComposeStepResult(ctx context.Context, sr *StepResult) context.Context {
	return WithLogger(ctx, GetLogger(ctx).WithFields(logrus.Fields{
		"StepIndex": sr.Index,
		"StepName":  sr.Name,
	}))
}

func GetLogger(ctx context.Context) logrus.FieldLogger {
	return ctx.Value(loggerKey).(logrus.FieldLogger)
}
