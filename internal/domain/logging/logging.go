package logging

import "context"

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
