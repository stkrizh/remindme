package verifyemailchannel

import (
	"context"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	CHANNEL_ID         = channel.ID(77)
	USER_ID            = user.ID(123)
	VERIFICATION_TOKEN = channel.VerificationToken("test")
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

func TestVerifyEmailChannelService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestSuccess() {
	s.ChannelRepository.GetByIDChannel.CreatedBy = USER_ID
	s.ChannelRepository.GetByIDChannel.Type = channel.Email
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		UserID:            USER_ID,
	})

	assert := s.Require()
	assert.Nil(err)
	assert.Equal(CHANNEL_ID, result.Channel.ID)
	assert.True(result.Channel.IsVerified())
	assert.Equal(Now, result.Channel.VerifiedAt.Value)
	assert.False(result.Channel.VerificationToken.IsPresent)
}

func (s *testSuite) TestChannelNotFoundByID() {
	s.ChannelRepository.GetByIDError = channel.ErrChannelDoesNotExist

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		UserID:            USER_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrChannelDoesNotExist)
	assert.False(result.Channel.IsVerified())
}

func (s *testSuite) TestChannelTypeIsNotEmail() {
	s.ChannelRepository.GetByIDChannel.CreatedBy = USER_ID
	s.ChannelRepository.GetByIDChannel.Type = channel.Telegram
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		UserID:            USER_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrInvalidVerificationData)
	assert.False(result.Channel.IsVerified())
}

func (s *testSuite) TestChannelBelongsToAnotherUser() {
	s.ChannelRepository.GetByIDChannel.CreatedBy = user.ID(100)
	s.ChannelRepository.GetByIDChannel.Type = channel.Email
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		UserID:            USER_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrInvalidVerificationData)
	assert.False(result.Channel.IsVerified())
}

func (s *testSuite) TestVerificationTokenIsInvalid() {
	s.ChannelRepository.GetByIDChannel.CreatedBy = USER_ID
	s.ChannelRepository.GetByIDChannel.Type = channel.Email
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: channel.VerificationToken("invalid-verification-token"),
		UserID:            USER_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrInvalidVerificationData)
	assert.False(result.Channel.IsVerified())
}

func (s *testSuite) TestChannelUpdateReturnsError() {
	s.ChannelRepository.GetByIDChannel.CreatedBy = USER_ID
	s.ChannelRepository.GetByIDChannel.Type = channel.Email
	s.ChannelRepository.GetByIDChannel.VerificationToken = common.NewOptional(VERIFICATION_TOKEN, true)
	s.ChannelRepository.UpdateError = channel.ErrChannelDoesNotExist

	result, err := s.Service.Run(context.Background(), Input{
		ChannelID:         CHANNEL_ID,
		VerificationToken: VERIFICATION_TOKEN,
		UserID:            USER_ID,
	})

	assert := s.Require()
	assert.ErrorIs(err, channel.ErrChannelDoesNotExist)
	assert.False(result.Channel.IsVerified())
}
