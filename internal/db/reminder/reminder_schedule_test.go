package remidner

import (
	"context"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/db"
	dbuser "remindme/internal/db/user"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

type testScheduleSuite struct {
	suite.Suite
	pool     *pgxpool.Pool
	repo     *PgxReminderRepository
	userRepo *dbuser.PgxUserRepository
	user     user.User
}

func (suite *testScheduleSuite) SetupSuite() {
	suite.pool = db.CreateTestPool()
	suite.repo = NewPgxReminderRepository(suite.pool)
	suite.userRepo = dbuser.NewPgxRepository(suite.pool)
}

func (suite *testScheduleSuite) TearDownSuite() {
	suite.pool.Close()
}

func (s *testScheduleSuite) SetupTest() {
	s.T().Helper()
	u, err := s.userRepo.Create(
		context.Background(),
		user.CreateUserInput{
			Email:        c.NewOptional(c.NewEmail("test-1@test.test"), true),
			PasswordHash: c.NewOptional(user.PasswordHash("test"), true),
			CreatedAt:    Now,
			ActivatedAt:  c.NewOptional(Now, true),
		},
	)
	s.Nil(err)
	s.user = u
}

func (suite *testScheduleSuite) TearDownTest() {
	db.TruncateTables(suite.pool)
}

func TestPgxReminderRepositorySchedule(t *testing.T) {
	suite.Run(t, new(testScheduleSuite))
}

func (s *testScheduleSuite) TestSchedule() {
	cases := []struct {
		id          string
		reminders   []reminder.CreateInput
		atBefore    time.Time
		now         time.Time
		expectedIds []int
	}{
		{
			id: "1",
			reminders: []reminder.CreateInput{
				{At: dt("2020-05-05T15:00:00Z"), Status: reminder.StatusCreated},
			},
			atBefore:    dt("2020-05-05T15:01:00Z"),
			now:         dt("2020-05-05T14:59:00Z"),
			expectedIds: []int{0},
		},
		{
			id: "2",
			reminders: []reminder.CreateInput{
				{At: dt("2020-05-05T15:00:00Z"), Status: reminder.StatusCreated},
			},
			atBefore:    dt("2020-05-05T14:59:00Z"),
			now:         dt("2020-05-05T14:49:00Z"),
			expectedIds: []int{},
		},
		{
			id: "3",
			reminders: []reminder.CreateInput{
				{At: dt("2020-05-05T15:01:00Z"), Status: reminder.StatusScheduled},
				{At: dt("2020-05-05T15:02:00Z"), Status: reminder.StatusSentSuccess},
				{At: dt("2020-05-05T15:03:00Z"), Status: reminder.StatusSentError},
				{At: dt("2020-05-05T15:04:00Z"), Status: reminder.StatusSentLimitExceeded},
				{At: dt("2020-05-05T15:05:00Z"), Status: reminder.StatusCanceled},
				{At: dt("2020-05-05T15:06:00Z"), Status: reminder.StatusCreated},
			},
			atBefore:    dt("2020-05-06T00:00:00Z"),
			now:         dt("2020-05-05T14:41:00Z"),
			expectedIds: []int{5},
		},
		{
			id: "4",
			reminders: []reminder.CreateInput{
				{At: dt("2020-05-05T15:01:00Z"), Status: reminder.StatusScheduled},
				{At: dt("2020-05-05T15:02:00Z"), Status: reminder.StatusCreated},
				{At: dt("2020-05-05T15:03:00Z"), Status: reminder.StatusSentError},
				{At: dt("2020-05-05T15:04:00Z"), Status: reminder.StatusCreated},
				{At: dt("2020-05-05T15:05:00Z"), Status: reminder.StatusCanceled},
				{At: dt("2020-05-05T15:06:00Z"), Status: reminder.StatusCreated},
			},
			atBefore:    dt("2020-05-05T15:05:00Z"),
			now:         dt("2020-05-05T14:41:00Z"),
			expectedIds: []int{1, 3},
		},
		{
			id: "5",
			reminders: []reminder.CreateInput{
				{At: dt("2020-05-05T15:01:00Z"), Status: reminder.StatusScheduled},
				{At: dt("2020-05-05T15:02:00Z"), Status: reminder.StatusCreated},
				{At: dt("2020-05-05T15:03:00Z"), Status: reminder.StatusSentError},
				{At: dt("2020-05-05T15:04:00Z"), Status: reminder.StatusCreated},
				{At: dt("2020-05-05T15:05:00Z"), Status: reminder.StatusCanceled},
				{At: dt("2020-05-05T15:06:00Z"), Status: reminder.StatusCreated},
			},
			atBefore:    dt("2020-05-05T15:07:00Z"),
			now:         dt("2020-05-05T14:41:00Z"),
			expectedIds: []int{1, 3, 5},
		},
	}

	for _, testcase := range cases {
		func() {
			defer s.truncateReminderTable()

			reminderIDs := s.createReminders(testcase.reminders)

			reminders, err := s.repo.Schedule(context.Background(), reminder.ScheduleInput{
				AtBefore:    testcase.atBefore,
				ScheduledAt: testcase.now,
			})

			s.Nil(err, testcase.id)
			s.assertScheduledCorrectly(testcase.id, reminderIDs, testcase.expectedIds, testcase.now, reminders)
		}()
	}
}

func (s *testScheduleSuite) truncateReminderTable() {
	s.T().Helper()

	_, err := s.pool.Exec(context.Background(), "DELETE FROM \"reminder\";")
	if err != nil {
		s.FailNow("could not truncate reminder table %w", err)
	}
}

func (s *testScheduleSuite) createReminders(inputs []reminder.CreateInput) []reminder.ID {
	s.T().Helper()

	reminderIDs := make([]reminder.ID, 0, len(inputs))
	for _, input := range inputs {
		input.CreatedBy = s.user.ID
		rem, err := s.repo.Create(context.Background(), input)

		assert := s.Require()
		assert.Nil(err)
		reminderIDs = append(reminderIDs, rem.ID)
	}

	return reminderIDs
}

func (s *testScheduleSuite) assertScheduledCorrectly(
	testcaseID string,
	reminderIDs []reminder.ID,
	expectedIndeces []int,
	expectedScheduledAt time.Time,
	actualReminders []reminder.Reminder,
) {
	s.T().Helper()

	expectedIDs := make([]reminder.ID, 0, len(expectedIndeces))
	for _, ix := range expectedIndeces {
		expectedIDs = append(expectedIDs, reminderIDs[ix])
	}

	actualIDs := make([]reminder.ID, 0, len(actualReminders))
	for _, rem := range actualReminders {
		s.Equal(rem.Status, reminder.StatusScheduled, testcaseID)
		s.Equal(rem.ScheduledAt, c.NewOptional(expectedScheduledAt, true), testcaseID)
		actualIDs = append(actualIDs, rem.ID)
	}

	s.ElementsMatch(expectedIDs, actualIDs, testcaseID)
}
