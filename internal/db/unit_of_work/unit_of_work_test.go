package uow

import (
	"context"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/db"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

const (
	REMINDER_ID_1 = 100
	REMINDER_ID_2 = 200
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

func TestPgxUnitOfWork(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestLimitsLock() {
	var wg sync.WaitGroup
	wg.Add(10)
	userID := s.createUserAndLimits()
	count := 0

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			uow, err := s.uow.Begin(ctx)
			if err != nil {
				return
			}
			defer uow.Rollback(ctx)

			_, err = uow.Limits().GetUserLimitsWithLock(ctx, userID)
			if err != nil {
				return
			}
			count += 1
		}()
	}

	wg.Wait()
	s.Equal(10, count)
}

func (s *testSuite) TestReminderLockOneReminder() {
	s.createReminders()

	var wg sync.WaitGroup
	wg.Add(10)
	count := 0

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			ctx := context.Background()
			uow, err := s.uow.Begin(ctx)
			if err != nil {
				s.Fail("could not begin unit of work")
				return
			}
			defer uow.Rollback(ctx)

			err = uow.Reminders().Lock(ctx, REMINDER_ID_1)
			c := count
			if err != nil {
				s.Fail("could not get lock by reminder ID, error is %v", err)
				return
			}

			_, err = uow.Reminders().GetByID(ctx, REMINDER_ID_1)
			if err != nil {
				s.Fail("could not get reminder by ID, error is %v", err)
				return
			}

			count = c + 1
		}()
	}

	wg.Wait()
	s.Equal(10, count)
}

func (s *testSuite) TestReminderLockTwoReminders() {
	s.createReminders()

	var wg sync.WaitGroup
	wg.Add(10)
	count_reminder_1 := 0
	count_reminder_2 := 0

	lockReminder := func(reminderID reminder.ID, count *int) {
		defer wg.Done()
		ctx := context.Background()
		uow, err := s.uow.Begin(ctx)
		if err != nil {
			s.Fail("could not begin unit of work")
			return
		}
		defer uow.Rollback(ctx)

		err = uow.Reminders().Lock(ctx, reminderID)
		c := *count
		if err != nil {
			s.Fail("could not get lock by reminder ID, error is %v", err)
			return
		}

		_, err = uow.Reminders().GetByID(ctx, reminderID)
		if err != nil {
			s.Fail("could not get reminder by ID, error is %v", err)
			return
		}

		*count = c + 1
	}

	for i := 0; i < 5; i++ {
		go func() {
			lockReminder(REMINDER_ID_1, &count_reminder_1)
			lockReminder(REMINDER_ID_2, &count_reminder_2)
		}()
	}

	wg.Wait()
	s.Equal(5, count_reminder_1)
	s.Equal(5, count_reminder_2)
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

func (s *testSuite) createReminders() {
	s.T().Helper()

	_, err := s.uow.db.Exec(
		context.Background(),
		`
		INSERT INTO "user" (id, email, password_hash, created_at) VALUES (1, 'test@test.test', 'test', now());
		`,
	)
	s.Nil(err)

	_, err = s.uow.db.Exec(
		context.Background(),
		`
		INSERT INTO channel (id, user_id, created_at, type, settings) VALUES
		(1, 1, now(), 'email', '{}'::jsonb),
		(2, 1, now(), 'email', '{}'::jsonb);
		`,
	)
	s.Nil(err)

	_, err = s.uow.db.Exec(
		context.Background(),
		`
		INSERT INTO reminder (id, user_id, at, created_at, status) VALUES 
		($1, 1, now(), now(), 'created'),
		($2, 1, now(), now(), 'created');
		`,
		REMINDER_ID_1,
		REMINDER_ID_2,
	)
	s.Nil(err)

	_, err = s.uow.db.Exec(
		context.Background(),
		`
		INSERT INTO reminder_channel (reminder_id, channel_id) VALUES ($1, 1), ($2, 1), ($2, 2);
		`,
		REMINDER_ID_1,
		REMINDER_ID_2,
	)
	s.Nil(err)
}
