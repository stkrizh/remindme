package logging

import (
	"context"
	"sync"
)

const DEBUG = "debug"
const INFO = "info"
const WARNING = "warning"
const ERROR = "error"

type FakeLoggerRecord struct {
	Level   string
	Msg     string
	Entries []LogEntry
}

type FakeLogger struct {
	Logged []FakeLoggerRecord
	lock   sync.RWMutex
}

func NewFakeLogger() *FakeLogger {
	return &FakeLogger{}
}

func (l *FakeLogger) Debug(ctx context.Context, msg string, entries ...LogEntry) {
	l.log(ctx, DEBUG, msg, entries...)
}

func (l *FakeLogger) Info(ctx context.Context, msg string, entries ...LogEntry) {
	l.log(ctx, INFO, msg, entries...)
}

func (l *FakeLogger) Warning(ctx context.Context, msg string, entries ...LogEntry) {
	l.log(ctx, WARNING, msg, entries...)
}

func (l *FakeLogger) Error(ctx context.Context, msg string, entries ...LogEntry) {
	l.log(ctx, ERROR, msg, entries...)
}

func (l *FakeLogger) log(ctx context.Context, level string, msg string, entries ...LogEntry) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.Logged = append(l.Logged, FakeLoggerRecord{
		Level:   DEBUG,
		Msg:     msg,
		Entries: entries,
	})
}
