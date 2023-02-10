package remindersender

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
)

type Sender struct {
	log logging.Logger
}

func New(log logging.Logger) *Sender {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	return &Sender{log: log}
}

func (s *Sender) SendReminder(ctx context.Context, rem reminder.ReminderWithChannels) error {
	s.log.Info(ctx, "Reminder has been sent.", logging.Entry("reminder", rem))
	return nil
}
