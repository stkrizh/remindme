package bot

import (
	"context"
	"remindme/internal/core/domain/channel"
)

type TelegramBotMessage struct {
	Bot    channel.TelegramBot
	ChatID channel.TelegramChatID
	Text   string
}

type TelegramBotMessageSender interface {
	SendTelegramBotMessage(ctx context.Context, m TelegramBotMessage) error
}
