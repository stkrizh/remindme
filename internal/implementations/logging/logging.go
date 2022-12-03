package logging

import (
	"context"
	"remindme/internal/domain/logging"

	"go.uber.org/zap"
)

type ZapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

func NewZapLogger() *ZapLogger {
	logger, err := zap.NewProduction(zap.AddCallerSkip(1))
	if err != nil {
		panic("Could not create Zap logger.")
	}
	sugar := logger.Sugar()
	return &ZapLogger{logger: logger, sugar: sugar}
}

func (l *ZapLogger) Sync() {
	l.logger.Sync()
}

func (l *ZapLogger) Debug(ctx context.Context, msg string, entries ...logging.LogEntry) {
	l.sugar.Debugf(msg, prepareArgs(entries...)...)
}

func (l *ZapLogger) Info(ctx context.Context, msg string, entries ...logging.LogEntry) {
	l.sugar.Infof(msg, prepareArgs(entries...)...)
}

func (l *ZapLogger) Warning(ctx context.Context, msg string, entries ...logging.LogEntry) {
	l.sugar.Warnf(msg, prepareArgs(entries...)...)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, entries ...logging.LogEntry) {
	l.sugar.Errorf(msg, prepareArgs(entries...)...)
}

func prepareArgs(entries ...logging.LogEntry) []interface{} {
	args := make([]interface{}, 0, len(entries)*2)
	for _, e := range entries {
		args = append(args, e.Key, e.Value)
	}
	return args
}
