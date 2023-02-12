package remindersender

import (
	"context"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/reminder"
)

type WebsocketSender struct{}

func NewWebsocket() *WebsocketSender {
	return &WebsocketSender{}
}

func (s *WebsocketSender) SendReminder(
	ctx context.Context,
	rem reminder.Reminder,
	settings *channel.WebsocketSettings,
) error {
	return nil
}
