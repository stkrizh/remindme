package activateuser

import (
	"context"
	"errors"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	EMAIL                  = "test@test.test"
	PASSWORD_HASH          = "test-password-hash"
	ACTIVATION_TOKEN       = "test-activation-token"
	INTERNAL_CHANNEL_TOKEN = "test-internal-channel-token"
)

var (
	Now           time.Time   = time.Now().UTC()
	DefaultLimits user.Limits = user.Limits{
		EmailChannelCount:    c.NewOptional(uint32(3), true),
		TelegramChannelCount: c.NewOptional(uint32(2), true),
	}
)

type testSuite struct {
	suite.Suite
	Logger  *logging.FakeLogger
	Uow     *uow.FakeUnitOfWork
	Service services.Service[Input, Result]
}

func (suite *testSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.Uow = uow.NewFakeUnitOfWork()
	suite.Service = New(
		suite.Logger,
		suite.Uow,
		func() time.Time { return Now },
		DefaultLimits,
	)
}

func TestActivateUserService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestSuccessUserActivated() {
	inactiveUser := s.createInactiveUser()

	_, err := s.Service.Run(
		context.Background(),
		Input{ActivationToken: user.ActivationToken(ACTIVATION_TOKEN)},
	)
	s.Nil(err)

	u, err := s.Uow.Context.UserRepository.GetByID(context.Background(), inactiveUser.ID)
	s.Nil(err)
	s.True(u.IsActive())

	s.True(s.Uow.Context.WasCommitCalled)
}

func (s *testSuite) TestSuccessLimitsCreated() {
	_ = s.createInactiveUser()

	_, err := s.Service.Run(
		context.Background(),
		Input{ActivationToken: user.ActivationToken(ACTIVATION_TOKEN)},
	)
	s.Nil(err)

	createdLimits := s.Uow.Context.LimitsRepository.Created
	s.Equal(1, len(createdLimits))
	s.Equal(DefaultLimits, createdLimits[0])
	s.True(s.Uow.Context.WasCommitCalled)
}

func (s *testSuite) TestSuccessEmailChannelCreated() {
	inactiveUser := s.createInactiveUser()

	_, err := s.Service.Run(
		context.Background(),
		Input{ActivationToken: user.ActivationToken(ACTIVATION_TOKEN)},
	)
	s.Nil(err)

	createdChannels := s.Uow.Context.ChannelRepository.Created
	s.Equal(2, len(createdChannels))
	createdEmailChannel := createdChannels[1]
	s.Equal(inactiveUser.ID, createdEmailChannel.CreatedBy)
	s.Equal(channel.Email, createdEmailChannel.Type)
	s.Equal(inactiveUser.Email.Value, createdEmailChannel.Settings.(*channel.EmailSettings).Email)
	s.True(createdEmailChannel.IsVerified())

	s.True(s.Uow.Context.WasCommitCalled)
}

func (s *testSuite) TestSuccessInternalChannelCreated() {
	inactiveUser := s.createInactiveUser()

	_, err := s.Service.Run(
		context.Background(),
		Input{ActivationToken: user.ActivationToken(ACTIVATION_TOKEN)},
	)
	s.Nil(err)

	createdChannels := s.Uow.Context.ChannelRepository.Created
	s.Equal(2, len(createdChannels))
	createdInternalChannel := createdChannels[0]
	s.Equal(inactiveUser.ID, createdInternalChannel.CreatedBy)
	s.Equal(channel.Internal, createdInternalChannel.Type)
	s.True(createdInternalChannel.IsVerified())

	s.True(s.Uow.Context.WasCommitCalled)
}

func (s *testSuite) TestChannelCreationFailed() {
	s.createInactiveUser()
	s.Uow.Context.ChannelRepository.CreateReturnsError = true

	_, err := s.Service.Run(
		context.Background(),
		Input{ActivationToken: user.ActivationToken(ACTIVATION_TOKEN)},
	)
	s.NotNil(err)
	s.False(s.Uow.Context.WasCommitCalled)
}

func (s *testSuite) TestActivationTokenInvalid() {
	activeUser := s.createInactiveUser()

	_, err := s.Service.Run(
		context.Background(),
		Input{ActivationToken: user.ActivationToken("invalid-token")},
	)
	s.True(errors.Is(err, user.ErrInvalidActivationToken))

	u, err := s.Uow.Context.UserRepository.GetByID(context.Background(), activeUser.ID)
	s.Nil(err)
	s.False(u.IsActive())
}

func (s *testSuite) createInactiveUser() user.User {
	s.T().Helper()
	u, err := s.Uow.Context.UserRepository.Create(
		context.Background(),
		user.CreateUserInput{
			Email:           c.NewOptional(c.NewEmail(EMAIL), true),
			PasswordHash:    c.NewOptional(user.PasswordHash(PASSWORD_HASH), true),
			CreatedAt:       Now,
			ActivationToken: c.NewOptional(user.ActivationToken(ACTIVATION_TOKEN), true),
		},
	)
	if err != nil {
		s.FailNow(err.Error())
	}
	s.False(u.IsActive())
	return u
}
