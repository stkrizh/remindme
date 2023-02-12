package reminder

import (
	"context"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
)

type Sender interface {
	SendReminder(ctx context.Context, reminder ReminderWithChannels) error
}

type EmailSender interface {
	SendReminder(ctx context.Context, reminder Reminder, settings *channel.EmailSettings) error
}

type TelegramSender interface {
	SendReminder(ctx context.Context, reminder Reminder, settings *channel.TelegramSettings) error
}

type WebsocketSender interface {
	SendReminder(ctx context.Context, reminder Reminder, settings *channel.WebsocketSettings) error
}

type ChannelSender struct {
	ctx             context.Context
	reminder        Reminder
	emailSender     EmailSender
	telegramSender  TelegramSender
	websocketSender WebsocketSender
}

func NewChannelSender(
	ctx context.Context,
	reminder Reminder,
	emailSender EmailSender,
	telegramSender TelegramSender,
	websocketSender WebsocketSender,
) *ChannelSender {
	if emailSender == nil {
		panic(e.NewNilArgumentError("emailSender"))
	}
	if telegramSender == nil {
		panic(e.NewNilArgumentError("telegramSender"))
	}
	if websocketSender == nil {
		panic(e.NewNilArgumentError("websocketSender"))
	}
	return &ChannelSender{
		ctx:             ctx,
		reminder:        reminder,
		emailSender:     emailSender,
		telegramSender:  telegramSender,
		websocketSender: websocketSender,
	}
}

func (s *ChannelSender) VisitEmail(settings *channel.EmailSettings) error {
	return s.emailSender.SendReminder(s.ctx, s.reminder, settings)
}

func (s *ChannelSender) VisitTelegram(settings *channel.TelegramSettings) error {
	return s.telegramSender.SendReminder(s.ctx, s.reminder, settings)
}

func (s *ChannelSender) VisitWebsocket(settings *channel.WebsocketSettings) error {
	return s.websocketSender.SendReminder(s.ctx, s.reminder, settings)
}

func (s *ChannelSender) SendReminder(settings channel.Settings) error {
	return settings.Accept(s)
}
