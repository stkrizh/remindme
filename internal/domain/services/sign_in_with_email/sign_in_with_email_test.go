package signinwithemail

import (
	"context"
	"errors"
	c "remindme/internal/domain/common"
	"remindme/internal/domain/logging"
	"remindme/internal/domain/user"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const EMAIL = "test@test.test"
const PASSWORD = "test-password"
const SESSION_TOKEN = "test-session-token"

var NOW time.Time = time.Now().UTC()

type testSuite struct {
	suite.Suite
	Logger                *logging.FakeLogger
	UserRepository        *user.FakeUserRepository
	SessionRepository     *user.FakeSessionRepository
	PasswordHasher        *user.FakePasswordHasher
	SessionTokenGenerator *user.FakeSessionTokenGenerator
	Service               *service
}

func (suite *testSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.UserRepository = user.NewFakeUserRepository()
	suite.SessionRepository = user.NewFakeSessionRepository(suite.UserRepository)
	suite.SessionTokenGenerator = user.NewFakeSessionTokenGenerator(SESSION_TOKEN)
	suite.Service = New(
		suite.Logger,
		suite.UserRepository,
		suite.SessionRepository,
		suite.PasswordHasher,
		suite.SessionTokenGenerator,
		func() time.Time { return NOW },
	)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestSuccess() {
	activeUser := s.createUser(true)

	result, err := s.Service.Run(
		context.Background(),
		Input{Email: user.NewEmail(EMAIL), Password: user.RawPassword(PASSWORD)},
	)

	s.Nil(err)
	u, err := s.SessionRepository.GetUserByToken(context.Background(), result.Token)
	s.Nil(err)
	s.Equal(activeUser, u)
}

func (s *testSuite) TestInvalidPassword() {
	s.createUser(true)

	result, err := s.Service.Run(
		context.Background(),
		Input{Email: user.NewEmail(EMAIL), Password: user.RawPassword("invalid-password")},
	)

	s.True(errors.Is(err, user.ErrInvalidCredentials))
	_, err = s.SessionRepository.GetUserByToken(context.Background(), result.Token)
	s.True(errors.Is(err, user.ErrUserDoesNotExist))
}

func (s *testSuite) TestInvalidEmail() {
	s.createUser(true)

	result, err := s.Service.Run(
		context.Background(),
		Input{Email: user.NewEmail(EMAIL + "test"), Password: user.RawPassword(PASSWORD)},
	)

	s.True(errors.Is(err, user.ErrInvalidCredentials))
	_, err = s.SessionRepository.GetUserByToken(context.Background(), result.Token)
	s.True(errors.Is(err, user.ErrUserDoesNotExist))
}

func (s *testSuite) TestUserIsNotActive() {
	s.createUser(false)

	result, err := s.Service.Run(
		context.Background(),
		Input{Email: user.NewEmail(EMAIL), Password: user.RawPassword(PASSWORD)},
	)

	s.True(errors.Is(err, user.ErrUserIsNotActive))
	_, err = s.SessionRepository.GetUserByToken(context.Background(), result.Token)
	s.True(errors.Is(err, user.ErrUserDoesNotExist))
}

func (s *testSuite) createUser(isActive bool) user.User {
	s.T().Helper()
	password, err := s.PasswordHasher.HashPassword(user.RawPassword(PASSWORD))
	if err != nil {
		s.FailNow(err.Error())
	}
	u, err := s.UserRepository.Create(
		context.Background(),
		user.CreateUserInput{
			Email:        c.NewOptional(user.NewEmail(EMAIL), true),
			PasswordHash: c.NewOptional(password, true),
			CreatedAt:    NOW,
			ActivatedAt:  c.NewOptional(NOW, isActive),
		},
	)
	if err != nil {
		s.FailNow(err.Error())
	}
	return u
}
