package user

import (
	"context"
	"remindme/internal/db"
	"remindme/internal/domain/common"
	"remindme/internal/domain/user"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

var now time.Time = time.Date(2020, 6, 6, 15, 30, 30, 0, time.UTC)

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

func (suite *testSuite) TestCreateSuccess() {
	type test struct {
		id    string
		input user.CreateUserInput
	}
	cases := []test{
		{
			id: "email",
			input: user.CreateUserInput{
				Email:           common.NewOptional(user.Email("test@test.test"), true),
				PasswordHash:    common.NewOptional(user.PasswordHash("test"), true),
				CreatedAt:       now,
				ActivationToken: common.NewOptional(user.ActivationToken("test"), true),
			},
		},
		{
			id: "identity",
			input: user.CreateUserInput{
				Identity:    common.NewOptional(user.Identity("test"), true),
				CreatedAt:   now,
				ActivatedAt: common.NewOptional(now, true),
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
			assert.False(u.LastLoginAt.IsPresent)
		})
	}

}

func (suite *testSuite) TestEmailAlreadyExistsError() {
	input := user.CreateUserInput{
		Email:           common.NewOptional(user.Email("test@test.test"), true),
		PasswordHash:    common.NewOptional(user.PasswordHash("test"), true),
		CreatedAt:       now,
		ActivationToken: common.NewOptional(user.ActivationToken("test"), true),
	}
	_, err := suite.repo.Create(context.Background(), input)

	assert := suite.Require()
	assert.Nil(err)

	_, err = suite.repo.Create(context.Background(), input)

	expectedErr := &user.EmailAlreadyExistsError{Email: user.Email("test@test.test")}
	assert.ErrorAs(err, &expectedErr)
}

func TestSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}
