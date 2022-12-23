package user

import (
	"context"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"remindme/internal/db"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

type testLimitsSuite struct {
	suite.Suite
	pool             *pgxpool.Pool
	userRepository   *PgxUserRepository
	limitsRepository *PgxLimitsRepository
}

func (suite *testLimitsSuite) SetupSuite() {
	suite.pool = db.CreateTestPool()
	suite.userRepository = NewPgxRepository(suite.pool)
	suite.limitsRepository = NewPgxLimitsRepository(suite.pool)
}

func (suite *testLimitsSuite) TearDownSuite() {
	suite.pool.Close()
}

func (suite *testLimitsSuite) TearDownTest() {
	db.TruncateTables(suite.pool)
}

func TestPgxLimitsRepository(t *testing.T) {
	suite.Run(t, new(testLimitsSuite))
}

func (s *testLimitsSuite) TestCreate() {
	cases := []struct {
		ID     string
		Limits user.Limits
	}{
		{ID: "1", Limits: user.Limits{}},
		{ID: "2", Limits: user.Limits{EmailChannelCount: c.NewOptional(uint32(0), true)}},
		{ID: "3", Limits: user.Limits{EmailChannelCount: c.NewOptional(uint32(1), true)}},
		{ID: "4", Limits: user.Limits{TelegramChannelCount: c.NewOptional(uint32(0), true)}},
		{ID: "5", Limits: user.Limits{TelegramChannelCount: c.NewOptional(uint32(1), true)}},
		{ID: "6", Limits: user.Limits{
			EmailChannelCount:    c.NewOptional(uint32(0), true),
			TelegramChannelCount: c.NewOptional(uint32(0), true),
		}},
		{ID: "7", Limits: user.Limits{
			EmailChannelCount:    c.NewOptional(uint32(1), true),
			TelegramChannelCount: c.NewOptional(uint32(1), true),
		}},
		{ID: "8", Limits: user.Limits{
			EmailChannelCount:    c.NewOptional(uint32(12345), true),
			TelegramChannelCount: c.NewOptional(uint32(12345), true),
		}},
	}

	for _, testCase := range cases {
		s.Run(testCase.ID, func() {
			defer db.TruncateTables(s.pool)
			activeUser := s.createActiveUser()
			limits, err := s.limitsRepository.Create(context.Background(), user.CreateLimitsInput{
				UserID: activeUser.ID,
				Limits: testCase.Limits,
			})
			s.Nil(err)
			s.Equal(testCase.Limits, limits)
		})
	}
}

func (s *testLimitsSuite) createActiveUser() user.User {
	s.T().Helper()
	u, err := s.userRepository.Create(
		context.Background(),
		user.CreateUserInput{
			Email:        c.NewOptional(c.NewEmail(EMAIL), true),
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
