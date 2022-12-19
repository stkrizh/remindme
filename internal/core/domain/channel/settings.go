package channel

import (
	c "remindme/internal/core/domain/common"
)

type Settings interface {
	Accept(visitor settingsVisitor) error
}

type settingsVisitor interface {
	VisitEmail(s *EmailSettings) error
	VisitTelegram(s *TelegramSettings) error
	VisitWebsocket(s *WebsocketSettings) error
}

type EmailSettings struct {
	Email c.Email
}

func NewEmailSettings(email c.Email) *EmailSettings {
	return &EmailSettings{Email: email}
}

func (s *EmailSettings) Accept(v settingsVisitor) error {
	return v.VisitEmail(s)
}

type TelegramBot string

type TelegramChatID int64

type TelegramSettings struct {
	Bot    TelegramBot
	ChatID TelegramChatID
}

func NewTelegramSettings(bot TelegramBot, chatID TelegramChatID) *TelegramSettings {
	return &TelegramSettings{Bot: bot, ChatID: chatID}
}

func (s *TelegramSettings) Accept(v settingsVisitor) error {
	return v.VisitTelegram(s)
}

type WebsocketSettings struct{}

func NewWebsocketSettings() *WebsocketSettings {
	return &WebsocketSettings{}
}

func (s *WebsocketSettings) Accept(v settingsVisitor) error {
	return v.VisitWebsocket(s)
}
