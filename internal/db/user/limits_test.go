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

func (s *testLimitsSuite) TestCreateAndGet() {
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
		{ID: "9", Limits: user.Limits{
			EmailChannelCount:        c.NewOptional(uint32(1), true),
			TelegramChannelCount:     c.NewOptional(uint32(1), true),
			ActiveReminderCount:      c.NewOptional(uint32(10), true),
			MonthlySentReminderCount: c.NewOptional(uint32(1000), true),
		}},
		{ID: "10", Limits: user.Limits{
			ActiveReminderCount:      c.NewOptional(uint32(5), true),
			MonthlySentReminderCount: c.NewOptional(uint32(50), true),
		}},
		{ID: "11", Limits: user.Limits{
			ReminderEveryPerDayCount: c.NewOptional(24.0, true),
		}},
		{ID: "12", Limits: user.Limits{
			EmailChannelCount:        c.NewOptional(uint32(1), true),
			TelegramChannelCount:     c.NewOptional(uint32(2), true),
			ActiveReminderCount:      c.NewOptional(uint32(10), true),
			MonthlySentReminderCount: c.NewOptional(uint32(1000), true),
			ReminderEveryPerDayCount: c.NewOptional(1.0, true),
		}},
	}

	for _, testCase := range cases {
		db.TruncateTables(s.pool)
		activeUser := s.createActiveUser()

		createdLimits, err := s.limitsRepository.Create(context.Background(), user.CreateLimitsInput{
			UserID: activeUser.ID,
			Limits: testCase.Limits,
		})
		s.Nil(err, testCase.ID)
		s.Equal(testCase.Limits, createdLimits, testCase.ID)

		readLimits, err := s.limitsRepository.GetUserLimits(context.Background(), activeUser.ID)
		s.Nil(err, testCase.ID)
		s.Equal(testCase.Limits, readLimits, testCase.ID)

		readLimits, err = s.limitsRepository.GetUserLimitsWithLock(context.Background(), activeUser.ID)
		s.Nil(err)
		s.Equal(testCase.Limits, readLimits, testCase.ID)
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
