package logging

import "context"

type entry struct {
	Key   string
	Value interface{}
}

func Entry(k string, v interface{}) entry {
	return entry{Key: k, Value: v}
}

type Logger interface {
	Debug(ctx context.Context, msg string, entries ...entry)
	Info(ctx context.Context, msg string, entries ...entry)
	Warning(ctx context.Context, msg string, entries ...entry)
	Error(ctx context.Context, msg string, entries ...entry)
}
