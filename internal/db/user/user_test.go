package user

import (
	"context"
	"errors"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"remindme/internal/db"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

const (
	EMAIL            = "test@test.test"
	PASSWORD_HASH    = "test-password-hash"
	ACTIVATION_TOKEN = "test-activation-token"
)

var NOW time.Time = time.Date(2020, 6, 6, 15, 30, 30, 0, time.UTC)

type testSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	repo *PgxUserRepository
}

func (suite *testSuite) SetupSuite() {
	suite.pool = db.CreateTestPool()
	suite.repo = NewPgxRepository(suite.pool)
}

func (suite *testSuite) TearDownSuite() {
	suite.pool.Close()
}

func (suite *testSuite) TearDownTest() {
	db.TruncateTables(suite.pool)
}

func TestPgxUserRepository(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (suite *testSuite) TestCreateSuccess() {
	type test struct {
		id    string
		input user.CreateUserInput
	}
	cases := []test{
		{
			id: "email",
			input: user.CreateUserInput{
				Email:           c.NewOptional(c.Email("test@test.test"), true),
				PasswordHash:    c.NewOptional(user.PasswordHash("test"), true),
				CreatedAt:       NOW,
				ActivationToken: c.NewOptional(user.ActivationToken("test"), true),
			},
		},
		{
			id: "identity",
			input: user.CreateUserInput{
				Identity:    c.NewOptional(user.Identity("test"), true),
				CreatedAt:   NOW,
				ActivatedAt: c.NewOptional(NOW, true),
			},
		},
	}

	for _, testcase := range cases {
		suite.Run(testcase.id, func() {
			u, err := suite.repo.Create(context.Background(), testcase.input)

			assert := suite.Require()
			assert.Nil(err)
			assert.Equal(testcase.input.Email, u.Email)
			assert.Equal(testcase.input.PasswordHash, u.PasswordHash)
			assert.Equal(testcase.input.Identity, u.Identity)
			assert.True(testcase.input.CreatedAt.Equal(u.CreatedAt))
			assert.Equal(testcase.input.ActivatedAt, u.ActivatedAt)
			assert.Equal(testcase.input.ActivationToken, u.ActivationToken)
		})
	}

}

func (suite *testSuite) TestEmailAlreadyExistsError() {
	input := user.CreateUserInput{
		Email:           c.NewOptional(c.Email("test@test.test"), true),
		PasswordHash:    c.NewOptional(user.PasswordHash("test"), true),
		CreatedAt:       NOW,
		ActivationToken: c.NewOptional(user.ActivationToken("test"), true),
	}
	_, err := suite.repo.Create(context.Background(), input)

	assert := suite.Require()
	assert.Nil(err)

	_, err = suite.repo.Create(context.Background(), input)
	assert.ErrorIs(err, user.ErrEmailAlreadyExists)
}

func (s *testSuite) TestActivateSuccess() {
	inactiveUser := s.createInactiveUser()
	activatedUser, err := s.repo.Activate(context.Background(), user.ActivationToken(ACTIVATION_TOKEN), NOW)

	s.Nil(err)
	s.Equal(inactiveUser.ID, activatedUser.ID)
	s.Equal(inactiveUser.Email, activatedUser.Email)
	s.Equal(inactiveUser.PasswordHash, activatedUser.PasswordHash)

	s.True(activatedUser.IsActive())
	s.True(activatedUser.ActivatedAt.IsPresent)
	s.Equal(NOW, activatedUser.ActivatedAt.Value)
}

func (s *testSuite) TestActivationFailsIfTokenIsInvalid() {
	inactiveUser := s.createInactiveUser()
	_, err := s.repo.Activate(context.Background(), user.ActivationToken("invalid-activate"), NOW)

	s.True(errors.Is(err, user.ErrUserDoesNotExist))

	u := s.getUserByID(inactiveUser.ID)
	s.False(u.IsActive())
}

func (s *testSuite) TestActivationFailsIfUserAlreadyActivated() {
	s.createInactiveUser()

	activatedUser, err := s.repo.Activate(context.Background(), user.ActivationToken(ACTIVATION_TOKEN), NOW)
	s.Nil(err)
	s.True(activatedUser.IsActive())

	_, err = s.repo.Activate(context.Background(), user.ActivationToken(ACTIVATION_TOKEN), NOW)
	s.True(errors.Is(err, user.ErrUserDoesNotExist))
}

func (s *testSuite) TestSetPassword() {
	u := s.createInactiveUser()
	s.True(u.PasswordHash.IsPresent)
	s.Equal(u.PasswordHash.Value, user.PasswordHash(PASSWORD_HASH))

	newPassword := user.PasswordHash("new-password-hash")
	err := s.repo.SetPassword(context.Background(), u.ID, newPassword)
	s.Nil(err)
	userAfterUpdate := s.getUserByID(u.ID)
	s.True(userAfterUpdate.PasswordHash.IsPresent)
	s.Equal(newPassword, userAfterUpdate.PasswordHash.Value)
}

func (s *testSuite) TestSetPasswordReturnsErrorIfUserDoesNotExist() {
	u := s.createInactiveUser()
	s.True(u.PasswordHash.IsPresent)
	s.Equal(u.PasswordHash.Value, user.PasswordHash(PASSWORD_HASH))

	newPassword := user.PasswordHash("new-password-hash")
	err := s.repo.SetPassword(context.Background(), user.ID(111222333), newPassword)
	s.True(errors.Is(err, user.ErrUserDoesNotExist))

	userAfterUpdate := s.getUserByID(u.ID)
	s.Equal(u, userAfterUpdate)
}

func (s *testSuite) createInactiveUser() user.User {
	s.T().Helper()
	u, err := s.repo.Create(
		context.Background(),
		user.CreateUserInput{
			Email:           c.NewOptional(c.NewEmail(EMAIL), true),
			PasswordHash:    c.NewOptional(user.PasswordHash(PASSWORD_HASH), true),
			CreatedAt:       NOW,
			ActivationToken: c.NewOptional(user.ActivationToken(ACTIVATION_TOKEN), true),
		},
	)
	if err != nil {
		s.FailNowf("could not create user", "err: %v", err)
	}
	s.False(u.IsActive())
	return u
}

func (s *testSuite) getUserByID(id user.ID) user.User {
	s.T().Helper()
	u, err := s.repo.GetByID(context.Background(), id)
	if err != nil {
		s.FailNowf("could not get user by ID", "id: %v, err: %v", id, err)
	}
	return u
}
