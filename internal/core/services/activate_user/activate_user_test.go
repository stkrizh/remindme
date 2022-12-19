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
	EMAIL            = "test@test.test"
	PASSWORD_HASH    = "test-password-hash"
	ACTIVATION_TOKEN = "test-activation-token"
)

var NOW time.Time = time.Now().UTC()

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
		func() time.Time { return NOW },
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

func (s *testSuite) TestSuccessEmailChannelCreated() {
	inactiveUser := s.createInactiveUser()

	_, err := s.Service.Run(
		context.Background(),
		Input{ActivationToken: user.ActivationToken(ACTIVATION_TOKEN)},
	)
	s.Nil(err)

	createdChannels := s.Uow.Context.ChannelRepository.Created
	s.Equal(1, len(createdChannels))
	createdChannel := createdChannels[0]
	s.Equal(inactiveUser.ID, createdChannel.CreatedBy)
	s.Equal(inactiveUser.Email.Value, createdChannel.Settings.(*channel.EmailSettings).Email)
	s.True(createdChannel.IsVerified())

	s.True(s.Uow.Context.WasCommitCalled)
}

func (s *testSuite) TestChannelCreationFailed() {
	s.createInactiveUser()
	s.Uow.Context.ChannelRepository.CreateReturnError = true

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
	s.True(errors.Is(err, user.ErrUserDoesNotExist))

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
			CreatedAt:       NOW,
			ActivationToken: c.NewOptional(user.ActivationToken(ACTIVATION_TOKEN), true),
		},
	)
	if err != nil {
		s.FailNow(err.Error())
	}
	s.False(u.IsActive())
	return u
}
