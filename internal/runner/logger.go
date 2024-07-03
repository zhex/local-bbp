package runner

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github/zhex/local-bbp/internal/common"
)

var loggerKey = "logger"

func NewLogger() logrus.FieldLogger {
	logger := logrus.StandardLogger()
	logger.SetLevel(logrus.GetLevel())
	logger.SetFormatter(&runnerLoggerFormatter{})
	return logger
}

func WithLogger(ctx context.Context, logger logrus.FieldLogger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}

func WithLoggerComposeStepResult(ctx context.Context, sr *StepResult) context.Context {
	return WithLogger(ctx, GetLogger(ctx).WithFields(logrus.Fields{
		"StepIndex": sr.GetIdxString(),
		"StepName":  sr.Name,
	}))
}

func GetLogger(ctx context.Context) logrus.FieldLogger {
	return ctx.Value(loggerKey).(logrus.FieldLogger)
}

type runnerLoggerFormatter struct {
	logrus.Formatter
}

func (f *runnerLoggerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}
	debug := ""
	if entry.Level == logrus.DebugLevel {
		debug = common.ColorCyan("[DEBUG] ")
	}
	stepInfo := ""
	if entry.Data["StepIndex"] != nil {
		stepInfo = fmt.Sprintf("[%s]", entry.Data["StepIndex"])
	}
	_, _ = fmt.Fprintf(b, "%s%s %s%s\n", common.ColorGrey(fmt.Sprintf("[%s]", entry.Data["ID"])), stepInfo, debug, entry.Message)
	return b.Bytes(), nil
}
