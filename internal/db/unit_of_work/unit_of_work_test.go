package uow

import (
	"context"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"remindme/internal/db"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	suite.Suite
	pool *pgxpool.Pool
	uow  *PgxUnitOfWork
}

func (suite *testSuite) SetupSuite() {
	suite.pool = db.CreateTestPool()
	suite.uow = NewPgxUnitOfWork(suite.pool)
}

func (suite *testSuite) TearDownSuite() {
	suite.pool.Close()
}

func (suite *testSuite) TearDownTest() {
	db.TruncateTables(suite.pool)
}

func TestPgxUnotOfWork(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestLimitsLock() {
	var wg sync.WaitGroup
	wg.Add(10)
	userID := s.createUserAndLimits()
	count := 0

	for i := 0; i < 10; i++ {
		go func() {
			ctx := context.Background()
			uow, err := s.uow.Begin(ctx)
			if err != nil {
				s.FailNowf("could not begin uow", "%w", err)
			}
			defer uow.Rollback(ctx)

			_, err = uow.Limits().GetUserLimitsWithLock(ctx, userID)
			if err != nil {
				s.FailNowf("could not get user limits", "%w", err)
			}
			count += 1
			err = uow.Commit(ctx)
			if err != nil {
				s.FailNowf("could not commit uow", "%w", err)
			}
			wg.Done()
		}()
	}

	wg.Wait()
	s.Equal(10, count)
}

func (s *testSuite) createUserAndLimits() user.ID {
	s.T().Helper()

	ctx := context.Background()
	uow, err := s.uow.Begin(ctx)
	if err != nil {
		s.FailNowf("could not begin uow", "%w", err)
	}
	defer uow.Rollback(ctx)

	createdUser, err := uow.Users().Create(ctx, user.CreateUserInput{
		Email:        c.NewOptional(c.NewEmail("test@test.com"), true),
		PasswordHash: c.NewOptional(user.PasswordHash("test"), true),
		CreatedAt:    time.Now().UTC(),
		ActivatedAt:  c.NewOptional(time.Now().UTC(), true),
	})
	if err != nil {
		s.FailNowf("could not create user", "%w", err)
	}

	_, err = uow.Limits().Create(ctx, user.CreateLimitsInput{
		UserID: createdUser.ID,
		Limits: user.Limits{
			EmailChannelCount:    c.NewOptional(uint32(1), true),
			TelegramChannelCount: c.NewOptional(uint32(2), true),
		},
	})
	if err != nil {
		s.FailNowf("could not create user limits", "%w", err)
	}

	uow.Commit(ctx)
	return createdUser.ID
}
