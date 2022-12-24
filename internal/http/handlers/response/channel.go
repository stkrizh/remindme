package response

import (
	"remindme/internal/core/domain/channel"
	"time"
)

type channelSettingsJSONEncoder struct {
	channel *Channel
}

func (e *channelSettingsJSONEncoder) VisitEmail(s *channel.EmailSettings) error {
	e.channel.EmailSettings = &EmailSettings{
		Email: string(s.Email),
	}
	return nil
}

func (e *channelSettingsJSONEncoder) VisitTelegram(s *channel.TelegramSettings) error {
	e.channel.TelegramSettings = &TelegramSettings{
		Bot:    string(s.Bot),
		ChatID: int64(s.ChatID),
	}
	return nil
}

func (e *channelSettingsJSONEncoder) VisitWebsocket(s *channel.WebsocketSettings) error {
	e.channel.WebsocketSettings = &WebsocketSettings{}
	return nil
}

type EmailSettings struct {
	Email string `json:"email"`
}

type TelegramSettings struct {
	Bot    string `json:"bot"`
	ChatID int64  `json:"chat_id"`
}

type WebsocketSettings struct{}

type Channel struct {
	ID                int64              `json:"id"`
	Type              string             `json:"type"`
	CreatedBy         int64              `json:"created_by"`
	CreatedAt         time.Time          `json:"created_at"`
	VerifiedAt        *time.Time         `json:"verified_at,omitempty"`
	EmailSettings     *EmailSettings     `json:"email,omitempty"`
	TelegramSettings  *TelegramSettings  `json:"telegram,omitempty"`
	WebsocketSettings *WebsocketSettings `json:"websocket,omitempty"`
}

func (c *Channel) FromDomainChannel(dc channel.Channel) {
	c.ID = int64(dc.ID)
	c.Type = dc.Type.String()
	c.CreatedBy = int64(dc.CreatedBy)
	c.CreatedAt = dc.CreatedAt
	if dc.VerifiedAt.IsPresent {
		c.VerifiedAt = &dc.VerifiedAt.Value
	}

	settingsEncoder := &channelSettingsJSONEncoder{channel: c}
	dc.Settings.Accept(settingsEncoder)
}