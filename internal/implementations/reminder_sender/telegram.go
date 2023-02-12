package remindersender

import (
	"context"
	"remindme/internal/core/domain/bot"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/reminder"
)

type TelegramSender struct {
	botMessageSender bot.TelegramBotMessageSender
}

func NewTelegram(botMessageSender bot.TelegramBotMessageSender) *TelegramSender {
	if botMessageSender == nil {
		panic(e.NewNilArgumentError("botMessageSender"))
	}
	return &TelegramSender{botMessageSender: botMessageSender}
}

func (s *TelegramSender) SendReminder(
	ctx context.Context,
	rem reminder.Reminder,
	settings *channel.TelegramSettings,
) error {
	text := "Hi there 👋 Let me remind you."
	if rem.Body != "" {
		text += "\n—\n" + rem.Body
	}
	return s.botMessageSender.SendTelegramBotMessage(
		ctx,
		bot.TelegramBotMessage{
			Bot:    settings.Bot,
			ChatID: settings.ChatID,
			Text:   text,
		},
	)
}
