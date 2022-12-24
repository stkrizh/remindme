package channel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/db/sqlcgen"
	"strconv"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

const (
	SETTINGS_EMAIL_EMAIL      = "email"
	SETTINGS_TELEGRAM_BOT     = "bot"
	SETTINGS_TELEGRAM_CHAT_ID = "chat_id"
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
	if input.Type == channel.Unknown {
		return c, fmt.Errorf("unknown channel type")
	}
	encodedSettings, err := encodeSettings(input.Settings)
	if err != nil {
		return c, err
	}
	dbChannel, err := r.queries.CreateChannel(
		ctx,
		sqlcgen.CreateChannelParams{
			UserID:    int64(input.CreatedBy),
			CreatedAt: input.CreatedAt,
			Type:      input.Type.String(),
			Settings:  encodedSettings,
			VerificationToken: sql.NullString{
				String: string(input.VerificationToken.Value),
				Valid:  input.VerificationToken.IsPresent,
			},
			VerifiedAt: sql.NullTime{
				Time:  input.VerifiedAt.Value,
				Valid: input.VerifiedAt.IsPresent,
			},
		},
	)
	if err != nil {
		return c, err
	}
	c, err = decodeChannel(dbChannel)
	return c, err
}

func (r *PgxChannelRepository) Read(
	ctx context.Context,
	options channel.ReadOptions,
) (channels []channel.Channel, err error) {
	dbChannels, err := r.queries.ReadChanels(
		ctx,
		sqlcgen.ReadChanelsParams{
			AllUserIds:   !options.UserIDEquals.IsPresent,
			UserIDEquals: int64(options.UserIDEquals.Value),
			AllTypes:     !options.TypeEquals.IsPresent,
			TypeEquals:   options.TypeEquals.Value.String(),
		},
	)
	if err != nil {
		return channels, err
	}
	channels = make([]channel.Channel, len(dbChannels))
	for ix, dbChannel := range dbChannels {
		channel, err := decodeChannel(dbChannel)
		if err != nil {
			return channels, err
		}
		channels[ix] = channel
	}
	return channels, nil
}

func (r *PgxChannelRepository) Count(
	ctx context.Context,
	options channel.ReadOptions,
) (count uint, err error) {
	rawCount, err := r.queries.CountChannels(
		ctx,
		sqlcgen.CountChannelsParams{
			AllUserIds:   !options.UserIDEquals.IsPresent,
			UserIDEquals: int64(options.UserIDEquals.Value),
			AllTypes:     !options.TypeEquals.IsPresent,
			TypeEquals:   options.TypeEquals.Value.String(),
		},
	)
	if err != nil {
		return count, err
	}
	return uint(rawCount), nil
}

func (r *PgxChannelRepository) Verify(
	ctx context.Context,
	input channel.VerifyInput,
) (c channel.Channel, err error) {
	dbChannel, err := r.queries.ActivateChannel(
		ctx,
		sqlcgen.ActivateChannelParams{
			ID:                int64(input.ID),
			UserID:            int64(input.CreatedBy),
			VerificationToken: string(input.VerificationToken),
			VerifiedAt:        input.At,
		},
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, channel.ErrChannelDoesNotExist
	}
	if err != nil {
		return c, err
	}
	domainChannel, err := decodeChannel(dbChannel)
	if err != nil {
		return c, err
	}
	return domainChannel, nil
}

func decodeChannel(dbChannel sqlcgen.Channel) (domainChannel channel.Channel, err error) {
	channelType := channel.ParseType(dbChannel.Type)
	if channelType == channel.Unknown {
		return domainChannel, fmt.Errorf("unknown channel (ID %d) type: %s", dbChannel.ID, dbChannel.Type)
	}
	settings, err := decodeSettings(channelType, dbChannel.Settings)
	if err != nil {
		return domainChannel, err
	}
	domainChannel = channel.Channel{
		ID:        channel.ID(dbChannel.ID),
		CreatedBy: user.ID(dbChannel.UserID),
		CreatedAt: dbChannel.CreatedAt,
		Type:      channelType,
		Settings:  settings,
		VerificationToken: c.NewOptional(
			channel.VerificationToken(dbChannel.VerificationToken.String),
			dbChannel.VerificationToken.Valid,
		),
		VerifiedAt: c.NewOptional(dbChannel.VerifiedAt.Time, dbChannel.VerifiedAt.Valid),
	}
	err = domainChannel.Validate()
	if err != nil {
		return domainChannel, err
	}
	return domainChannel, nil
}

type settingsJSONBEncoder struct {
	result pgtype.JSONB
}

func (c *settingsJSONBEncoder) VisitEmail(s *channel.EmailSettings) error {
	settings := make(map[string]interface{})
	settings[SETTINGS_EMAIL_EMAIL] = string(s.Email)
	if err := c.result.Set(settings); err != nil {
		return err
	}
	return nil
}

func (c *settingsJSONBEncoder) VisitTelegram(s *channel.TelegramSettings) error {
	settings := make(map[string]interface{})
	settings[SETTINGS_TELEGRAM_BOT] = string(s.Bot)
	settings[SETTINGS_TELEGRAM_CHAT_ID] = fmt.Sprintf("%d", s.ChatID)
	if err := c.result.Set(settings); err != nil {
		return err
	}
	return nil
}

func (c *settingsJSONBEncoder) VisitWebsocket(s *channel.WebsocketSettings) error {
	settings := make(map[string]interface{})
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

func (d *settingsJSONBDecoder) VisitEmail(s *channel.EmailSettings) error {
	rawEmail, ok := d.encoded[SETTINGS_EMAIL_EMAIL]
	if !ok {
		return fmt.Errorf("could not get email from channel settings: %v", d.encoded)
	}
	email, ok := rawEmail.(string)
	if !ok {
		return fmt.Errorf("email is not a string: %v", d.encoded)
	}
	s.Email = c.NewEmail(email)
	return nil
}

func (d *settingsJSONBDecoder) VisitTelegram(s *channel.TelegramSettings) error {
	rawBot, ok := d.encoded[SETTINGS_TELEGRAM_BOT]
	if !ok {
		return fmt.Errorf("could not get telegram bot from channel settings: %v", d.encoded)
	}
	bot, ok := rawBot.(string)
	if !ok {
		return fmt.Errorf("bot is not a string: %v", d.encoded)
	}
	s.Bot = channel.TelegramBot(bot)

	rawChatID, ok := d.encoded[SETTINGS_TELEGRAM_CHAT_ID]
	if !ok {
		return fmt.Errorf("could not get telegram chat ID from channel settings: %v", d.encoded)
	}
	strChatID, ok := rawChatID.(string)
	if !ok {
		return fmt.Errorf("invalid telegram chat ID: %v", d.encoded)
	}
	chatID, err := strconv.ParseInt(strChatID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid telegram chat ID: %w, %v", err, d.encoded)
	}
	s.ChatID = channel.TelegramChatID(chatID)

	return nil
}

func (d *settingsJSONBDecoder) VisitWebsocket(s *channel.WebsocketSettings) error {
	return nil
}

func decodeSettings(channelType channel.Type, encoded pgtype.JSONB) (settings channel.Settings, err error) {
	m, ok := encoded.Get().(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("could not cast JSONB encoded value: %v", encoded)
	}

	switch channelType {
	case channel.Email:
		settings = &channel.EmailSettings{}
	case channel.Telegram:
		settings = &channel.TelegramSettings{}
	case channel.Websocket:
		settings = &channel.WebsocketSettings{}
	default:
		return nil, fmt.Errorf("unknown channel settings type: %v", m)
	}

	decoder := &settingsJSONBDecoder{encoded: m}
	err = settings.Accept(decoder)
	return settings, err
}
