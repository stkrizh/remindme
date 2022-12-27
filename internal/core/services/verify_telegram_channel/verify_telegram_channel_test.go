package verifytelegramchannel

import (
	"context"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/services"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	CHANNEL_ID         = channel.ID(77)
	VERIFICATION_TOKEN = channel.VerificationToken("test")
	TELEGRAM_BOT       = channel.TelegramBot("test")
	TELEGRAM_CHAT_ID   = channel.TelegramChatID(123)
)

var Now time.Time = time.Now().UTC()

type testSuite struct {
	suite.Suite
	Logger            *logging.FakeLogger
	ChannelRepository *channel.FakeRepository
	Service           services.Service[Input, Result]
}

func (suite *testSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.ChannelRepository = channel.NewFakeRepository()
	suite.Service = New(
		suite.Logger,
		suite.ChannelRepository,
		func() time.Time { return Now },
	)
}

func TestVerifyTelegramChannelService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestSuccess() {
	s.ChannelRepository.GetByIDChannel.Type = channel.Telegram
	s.ChannelRepository.GetByIDChannel.Settings = channel.NewTelegramSettings(TELEGRAM_BOT, channel.TelegramChatID(0))
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		TelegramBot:       TELEGRAM_BOT,
		TelegramChatID:    TELEGRAM_CHAT_ID,
	})

	assert := s.Require()
	assert.Nil(err)
	assert.Equal(CHANNEL_ID, result.Channel.ID)
	assert.True(result.Channel.IsVerified())
	assert.Equal(Now, result.Channel.VerifiedAt.Value)
	assert.False(result.Channel.VerificationToken.IsPresent)
	assert.Equal(channel.NewTelegramSettings(TELEGRAM_BOT, TELEGRAM_CHAT_ID), result.Channel.Settings)
}

func (s *testSuite) TestChannelNotFoundByID() {
	s.ChannelRepository.GetByIDError = channel.ErrChannelDoesNotExist

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		TelegramBot:       TELEGRAM_BOT,
		TelegramChatID:    TELEGRAM_CHAT_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrChannelDoesNotExist)
	assert.False(result.Channel.IsVerified())
}

func (s *testSuite) TestChannelTypeIsNotTelegram() {
	s.ChannelRepository.GetByIDChannel.Type = channel.Email
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		TelegramBot:       TELEGRAM_BOT,
		TelegramChatID:    TELEGRAM_CHAT_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrInvalidVerificationData)
	assert.False(result.Channel.IsVerified())
}

func (s *testSuite) TestInvalidTelegramBot() {
	s.ChannelRepository.GetByIDChannel.Type = channel.Telegram
	s.ChannelRepository.GetByIDChannel.Settings = channel.NewTelegramSettings(TELEGRAM_BOT, channel.TelegramChatID(0))
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		TelegramBot:       channel.TelegramBot("invalid-telegram-bot"),
		TelegramChatID:    TELEGRAM_CHAT_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrInvalidVerificationData)
	assert.False(result.Channel.IsVerified())
}

func (s *testSuite) TestVerificationTokenIsInvalid() {
	s.ChannelRepository.GetByIDChannel.Type = channel.Telegram
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: channel.VerificationToken("invalid-verification-token"),
		TelegramBot:       TELEGRAM_BOT,
		TelegramChatID:    TELEGRAM_CHAT_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrInvalidVerificationData)
	assert.False(result.Channel.IsVerified())
}

func (s *testSuite) TestChannelUpdateReturnsError() {
	s.ChannelRepository.GetByIDChannel.Type = channel.Telegram
	s.ChannelRepository.GetByIDChannel.Settings = channel.NewTelegramSettings(TELEGRAM_BOT, channel.TelegramChatID(0))
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)
	s.ChannelRepository.UpdateError = channel.ErrChannelDoesNotExist

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		TelegramBot:       TELEGRAM_BOT,
		TelegramChatID:    TELEGRAM_CHAT_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrChannelDoesNotExist)
	assert.False(result.Channel.IsVerified())
}
