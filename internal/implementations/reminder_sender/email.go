package remindersender

import (
	"context"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/reminder"
)

type EmailSender struct{}

func NewEmail() *EmailSender {
	return &EmailSender{}
}

func (s *EmailSender) SendReminder(
	ctx context.Context,
	rem reminder.Reminder,
	settings *channel.EmailSettings,
) error {
	return nil
}
