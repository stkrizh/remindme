package channel

import (
	c "remindme/internal/core/domain/common"
)

type Settings interface {
	Accept(visitor SettingsVisitor) error
}

type SettingsVisitor interface {
	VisitEmail(s *EmailSettings) error
	VisitTelegram(s *TelegramSettings) error
	VisitInternal(s *InternalSettings) error
}

type EmailSettings struct {
	Email c.Email
}

func NewEmailSettings(email c.Email) *EmailSettings {
	return &EmailSettings{Email: email}
}

func (s *EmailSettings) Accept(v SettingsVisitor) error {
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

func (s *TelegramSettings) Accept(v SettingsVisitor) error {
	return v.VisitTelegram(s)
}

type InternalChannelToken string

type InternalSettings struct {
	Token InternalChannelToken
}

func NewInternalSettings(token InternalChannelToken) *InternalSettings {
	return &InternalSettings{
		Token: token,
	}
}

func (s *InternalSettings) Accept(v SettingsVisitor) error {
	return v.VisitInternal(s)
}

type InternalChannelTokenGenerator interface {
	GenerateInternalChannelToken() InternalChannelToken
}

type InternalChannelTokenValidator interface {
	ValidateInternalChannelToken(token InternalChannelToken) bool
}
