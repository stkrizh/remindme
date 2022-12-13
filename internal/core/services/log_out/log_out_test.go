package logout

import (
	"context"
	"errors"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	EMAIL         = "test@test.test"
	PASSWORD_HASH = "test-password-hash"
	SESSION_TOKEN = "test-session-token"
)

var NOW time.Time = time.Now().UTC()

type testSuite struct {
	suite.Suite
	Logger            *logging.FakeLogger
	UserRepository    *user.FakeUserRepository
	SessionRepository *user.FakeSessionRepository
	Service           services.Service[Input, Result]
}

func (suite *testSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.UserRepository = user.NewFakeUserRepository()
	suite.SessionRepository = user.NewFakeSessionRepository(suite.UserRepository)
	suite.Service = New(
		suite.Logger,
		suite.SessionRepository,
	)
}

func TestLogOutService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestSuccess() {
	s.createUserAndSession()

	_, err := s.Service.Run(
		context.Background(),
		Input{Token: user.SessionToken(SESSION_TOKEN)},
	)
	s.Nil(err)
	s.False(s.sessionExists(user.SessionToken(SESSION_TOKEN)))
}

func (s *testSuite) TestErrorReturnedIfSessionTokenInvalid() {
	s.createUserAndSession()

	_, err := s.Service.Run(
		context.Background(),
		Input{Token: user.SessionToken("invalid-session-token")},
	)
	s.True(errors.Is(err, user.ErrSessionDoesNotExist))
	s.True(s.sessionExists(user.SessionToken(SESSION_TOKEN)))
}

func (s *testSuite) createUserAndSession() user.User {
	s.T().Helper()
	u, err := s.UserRepository.Create(
		context.Background(),
		user.CreateUserInput{
			Email:        c.NewOptional(user.NewEmail(EMAIL), true),
			PasswordHash: c.NewOptional(user.PasswordHash(PASSWORD_HASH), true),
			CreatedAt:    NOW,
			ActivatedAt:  c.NewOptional(NOW, true),
		},
	)
	if err != nil {
		s.FailNow(err.Error())
	}
	s.True(u.IsActive())

	err = s.SessionRepository.Create(
		context.Background(),
		user.CreateSessionInput{
			UserID:    u.ID,
			Token:     user.SessionToken(SESSION_TOKEN),
			CreatedAt: NOW,
		},
	)
	if err != nil {
		s.FailNow(err.Error())
	}
	s.True(s.sessionExists(user.SessionToken(SESSION_TOKEN)))
	return u
}

func (s *testSuite) sessionExists(token user.SessionToken) bool {
	s.T().Helper()
	_, err := s.SessionRepository.GetUserByToken(context.Background(), user.SessionToken(SESSION_TOKEN))
	if errors.Is(err, user.ErrUserDoesNotExist) {
		return false
	}
	if err != nil {
		s.FailNow(err.Error())
	}
	return true
}
