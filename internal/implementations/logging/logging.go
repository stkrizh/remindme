package logging

import (
	"context"
	"fmt"
	"remindme/internal/core/domain/logging"

	"github.com/getsentry/sentry-go"
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
	l.sugar.Debugw(msg, prepareArgs(entries...)...)
}

func (l *ZapLogger) Info(ctx context.Context, msg string, entries ...logging.LogEntry) {
	l.sugar.Infow(msg, prepareArgs(entries...)...)
}

func (l *ZapLogger) Warning(ctx context.Context, msg string, entries ...logging.LogEntry) {
	l.sugar.Warnw(msg, prepareArgs(entries...)...)
}

func (l *ZapLogger) Error(ctx context.Context, msg string, entries ...logging.LogEntry) {
	l.sugar.Errorw(msg, prepareArgs(entries...)...)
	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
	}
	if hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			e := make(map[string]interface{}, len(entries))
			for _, entry := range entries {
				e[entry.Key] = fmt.Sprintf("%+v", entry.Value)
			}
			scope.SetContext("entries", e)
			hub.CaptureMessage(msg)
		})
	}
}

func prepareArgs(entries ...logging.LogEntry) []interface{} {
	args := make([]interface{}, 0, len(entries)*2)
	for _, e := range entries {
		args = append(args, e.Key, e.Value)
	}
	return args
}
