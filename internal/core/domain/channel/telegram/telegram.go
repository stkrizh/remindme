package telegram

import "remindme/internal/core/domain/channel"

type BotToken string

type ChatId int64

type Settings struct {
	BotToken BotToken
	ChatId   ChatId
}

func (s Settings) Type() channel.Type {
	return channel.TELEGRAM
}
