package user

import (
	"context"
	"errors"
	"remindme/internal/db"
	c "remindme/internal/domain/common"
	"remindme/internal/domain/user"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

const (
	SESSION_TOKEN = "test-session-token"
)

type testSessionSuite struct {
	suite.Suite
	pool              *pgxpool.Pool
	userRepository    *PgxUserRepository
	sessionRepository *PgxSessionRepository
}

func (suite *testSessionSuite) SetupSuite() {
	suite.pool = db.CreateTestPool()
	suite.userRepository = NewPgxRepository(suite.pool)
	suite.sessionRepository = NewPgxSessionRepository(suite.pool)
}

func (suite *testSessionSuite) TearDownSuite() {
	suite.pool.Close()
}

func (suite *testSessionSuite) TearDownTest() {
	db.TruncateTables(suite.pool)
}

func TestPgxSessionRepository(t *testing.T) {
	suite.Run(t, new(testSessionSuite))
}

func (s *testSessionSuite) TestCreate() {
	activeUser := s.createActiveUser()

	err := s.sessionRepository.Create(
		context.Background(),
		user.CreateSessionInput{
			UserID:    activeUser.ID,
			Token:     user.SessionToken(SESSION_TOKEN),
			CreatedAt: NOW,
		},
	)
	u, ok := s.getUserByToken(user.SessionToken(SESSION_TOKEN))
	s.Nil(err)
	s.True(ok)
	s.Equal(activeUser.ID, u.ID)
}

func (s *testSessionSuite) TestDeleteSuccess() {
	activeUser := s.createActiveUser()

	err := s.sessionRepository.Create(
		context.Background(),
		user.CreateSessionInput{
			UserID:    activeUser.ID,
			Token:     user.SessionToken(SESSION_TOKEN),
			CreatedAt: NOW,
		},
	)
	s.Nil(err)

	userID, err := s.sessionRepository.Delete(context.Background(), user.SessionToken(SESSION_TOKEN))
	s.Nil(err)
	s.Equal(activeUser.ID, userID)
	_, ok := s.getUserByToken(user.SessionToken(SESSION_TOKEN))
	s.False(ok)
}

func (s *testSessionSuite) createActiveUser() user.User {
	s.T().Helper()
	u, err := s.userRepository.Create(
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
	return u
}

func (s *testSessionSuite) getUserByToken(token user.SessionToken) (u user.User, ok bool) {
	u, err := s.sessionRepository.GetUserByToken(context.Background(), token)
	if errors.Is(err, user.ErrUserDoesNotExist) {
		return u, false
	}
	if err != nil {
		s.FailNow(err.Error())
	}
	return u, true
}
