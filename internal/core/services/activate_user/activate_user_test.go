package activateuser

import (
	"context"
	"errors"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
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
	Logger         *logging.FakeLogger
	UserRepository *user.FakeUserRepository
	Service        *service
}

func (suite *testSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.UserRepository = user.NewFakeUserRepository()
	suite.Service = New(
		suite.Logger,
		suite.UserRepository,
		func() time.Time { return NOW },
	)
}

func TestActivateUserService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestSuccess() {
	activeUser := s.createInactiveUser()

	_, err := s.Service.Run(
		context.Background(),
		Input{ActivationToken: user.ActivationToken(ACTIVATION_TOKEN)},
	)
	s.Nil(err)

	u, err := s.UserRepository.GetByID(context.Background(), activeUser.ID)
	s.Nil(err)
	s.True(u.IsActive())
}

func (s *testSuite) TestActivationTokenInvalid() {
	activeUser := s.createInactiveUser()

	_, err := s.Service.Run(
		context.Background(),
		Input{ActivationToken: user.ActivationToken("invalid-token")},
	)
	s.True(errors.Is(err, user.ErrUserDoesNotExist))

	u, err := s.UserRepository.GetByID(context.Background(), activeUser.ID)
	s.Nil(err)
	s.False(u.IsActive())
}

func (s *testSuite) createInactiveUser() user.User {
	s.T().Helper()
	u, err := s.UserRepository.Create(
		context.Background(),
		user.CreateUserInput{
			Email:           c.NewOptional(user.NewEmail(EMAIL), true),
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
