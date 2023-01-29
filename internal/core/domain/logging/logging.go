package logging

import (
	"context"
	"errors"
)

type LogEntry struct {
	Key   string
	Value interface{}
}

func Entry(k string, v interface{}) LogEntry {
	return LogEntry{Key: k, Value: v}
}

type Logger interface {
	Debug(ctx context.Context, msg string, entries ...LogEntry)
	Info(ctx context.Context, msg string, entries ...LogEntry)
	Warning(ctx context.Context, msg string, entries ...LogEntry)
	Error(ctx context.Context, msg string, entries ...LogEntry)
}

func Error(ctx context.Context, logger Logger, err error, entries ...LogEntry) {
	if errors.Is(err, context.Canceled) {
		return
	}
	entries = append(entries, Entry("err", err))
	logger.Error(ctx, "Unexpected error occured.", entries...)
}
