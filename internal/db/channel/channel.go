package channel

import (
	"context"
	"fmt"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/db/sqlcgen"
	"strconv"

	"github.com/jackc/pgtype"
)

const (
	SETTINGS_TYPE_FIELD         = "type"
	SETTINGS_EMAIL              = "email"
	SETTINGS_EMAIL_EMAIL        = "email"
	SETTINGS_TELEGRAM           = "telegram"
	SETTINGS_TELEGRAM_BOT_TOKEN = "bot_token"
	SETTINGS_TELEGRAM_CHAT_ID   = "chat_id"
	SETTINGS_WEBSOCKET          = "websocket"
)

type PgxChannelRepository struct {
	queries *sqlcgen.Queries
}

func NewPgxChannelRepository(db sqlcgen.DBTX) *PgxChannelRepository {
	if db == nil {
		panic(e.NewNilArgumentError("db"))
	}
	return &PgxChannelRepository{queries: sqlcgen.New(db)}
}

func (r *PgxChannelRepository) Create(ctx context.Context, input channel.CreateInput) (c channel.Channel, err error) {
	encodedSettings, err := encodeSettings(input.Settings)
	if err != nil {
		return c, err
	}
	dbChannel, err := r.queries.CreateChannel(
		ctx,
		sqlcgen.CreateChannelParams{
			UserID:     int64(input.CreatedBy),
			CreatedAt:  input.CreatedAt,
			Settings:   encodedSettings,
			IsVerified: input.IsVerified,
		},
	)
	if err != nil {
		return c, err
	}
	c, err = decodeChannel(dbChannel)
	return c, err
}

func decodeChannel(c sqlcgen.Channel) (dc channel.Channel, err error) {
	settings, err := decodeSettings(c.Settings)
	if err != nil {
		return dc, err
	}
	return channel.Channel{
		ID:         channel.ID(c.ID),
		CreatedBy:  user.ID(c.UserID),
		CreatedAt:  c.CreatedAt,
		Settings:   settings,
		IsVerified: c.IsVerified,
	}, nil
}

type settingsJSONBEncoder struct {
	result pgtype.JSONB
}

func (c *settingsJSONBEncoder) VisitEmail(s *channel.EmailSettings) error {
	settings := make(map[string]interface{})
	settings[SETTINGS_TYPE_FIELD] = SETTINGS_EMAIL
	settings[SETTINGS_EMAIL_EMAIL] = string(s.Email)
	if err := c.result.Set(settings); err != nil {
		return err
	}
	return nil
}

func (c *settingsJSONBEncoder) VisitTelegram(s *channel.TelegramSettings) error {
	settings := make(map[string]interface{})
	settings[SETTINGS_TYPE_FIELD] = SETTINGS_TELEGRAM
	settings[SETTINGS_TELEGRAM_BOT_TOKEN] = string(s.BotToken)
	settings[SETTINGS_TELEGRAM_CHAT_ID] = fmt.Sprintf("%d", s.ChatID)
	if err := c.result.Set(settings); err != nil {
		return err
	}
	return nil
}

func (c *settingsJSONBEncoder) VisitWebsocket(s *channel.WebsocketSettings) error {
	settings := make(map[string]interface{})
	settings[SETTINGS_TYPE_FIELD] = SETTINGS_WEBSOCKET
	if err := c.result.Set(settings); err != nil {
		return err
	}
	return nil
}

func encodeSettings(settings channel.Settings) (encoded pgtype.JSONB, err error) {
	settingsEncoder := &settingsJSONBEncoder{}
	err = settings.Accept(settingsEncoder)
	if err != nil {
		return encoded, fmt.Errorf("could not encode channel settings due to error: %w", err)
	}
	return settingsEncoder.result, nil
}

type settingsJSONBDecoder struct {
	encoded map[string]interface{}
}

func (c *settingsJSONBDecoder) VisitEmail(s *channel.EmailSettings) error {
	rawEmail, ok := c.encoded[SETTINGS_EMAIL_EMAIL]
	if !ok {
		return fmt.Errorf("could not get email from channel settings: %v", c.encoded)
	}
	email, ok := rawEmail.(string)
	if !ok {
		return fmt.Errorf("email is not a string: %v", c.encoded)
	}
	s.Email = common.NewEmail(email)
	return nil
}

func (c *settingsJSONBDecoder) VisitTelegram(s *channel.TelegramSettings) error {
	rawBotToken, ok := c.encoded[SETTINGS_TELEGRAM_BOT_TOKEN]
	if !ok {
		return fmt.Errorf("could not get telegram bot token from channel settings: %v", c.encoded)
	}
	botToken, ok := rawBotToken.(string)
	if !ok {
		return fmt.Errorf("bot token is not a string: %v", c.encoded)
	}
	s.BotToken = channel.TelegramBotToken(botToken)

	rawChatID, ok := c.encoded[SETTINGS_TELEGRAM_CHAT_ID]
	if !ok {
		return fmt.Errorf("could not get telegram chat ID from channel settings: %v", c.encoded)
	}
	strChatID, ok := rawChatID.(string)
	if !ok {
		return fmt.Errorf("invalid telegram chat ID: %v", c.encoded)
	}
	chatID, err := strconv.ParseInt(strChatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat ID: %w, %v", err, c.encoded)
	}
	s.ChatID = channel.TelegramChatID(chatID)

	return nil
}

func (c *settingsJSONBDecoder) VisitWebsocket(s *channel.WebsocketSettings) error {
	return nil
}

func decodeSettings(encoded pgtype.JSONB) (settings channel.Settings, err error) {
	m, ok := encoded.Get().(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not cast JSONB encoded value: %v", encoded)
	}
	settingsType, ok := m[SETTINGS_TYPE_FIELD]
	if !ok {
		return nil, fmt.Errorf("could not define channel settings type: %v", m)
	}
	switch settingsType {
	case SETTINGS_EMAIL:
		settings = &channel.EmailSettings{}
	case SETTINGS_TELEGRAM:
		settings = &channel.TelegramSettings{}
	case SETTINGS_WEBSOCKET:
		settings = &channel.WebsocketSettings{}
	default:
		return nil, fmt.Errorf("unknown channel settings type: %v", m)
	}

	decoder := &settingsJSONBDecoder{encoded: m}
	err = settings.Accept(decoder)
	return settings, err
}
