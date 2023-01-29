package schedulereminders

import (
	"context"
	"errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

var (
	Now = time.Now().UTC()
)

type testSuite struct {
	suite.Suite
	logger     *logging.FakeLogger
	unitOfWork *uow.FakeUnitOfWork
	scheduler  *reminder.TestReminderScheduler
	service    services.Service[Input, Result]
}

func (suite *testSuite) SetupTest() {
	suite.logger = logging.NewFakeLogger()
	suite.unitOfWork = uow.NewFakeUnitOfWork()
	suite.scheduler = reminder.NewTestReminderScheduler()
	suite.service = New(
		suite.logger,
		suite.unitOfWork,
		suite.scheduler,
		func() time.Time { return Now },
	)
}

func (suite *testSuite) TearDownTest() {
	suite.scheduler.Scheduled = make([]reminder.Reminder, 0)
}

func TestScheduleRemindersService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestSuccess() {
	cases := []struct {
		id        string
		now       time.Time
		reminders []reminder.Reminder
	}{
		{
			id:        "1",
			now:       Now,
			reminders: []reminder.Reminder{},
		},
		{
			id:  "2",
			now: Now,
			reminders: []reminder.Reminder{
				{ID: reminder.ID(1), CreatedBy: user.ID(100), Status: reminder.StatusCreated},
				{ID: reminder.ID(2), CreatedBy: user.ID(200), Status: reminder.StatusCreated},
				{ID: reminder.ID(3), CreatedBy: user.ID(200), Status: reminder.StatusCreated},
			},
		},
		{
			id:  "3",
			now: time.Date(2020, 1, 1, 15, 0, 0, 0, time.UTC),
			reminders: []reminder.Reminder{
				{ID: reminder.ID(1), CreatedBy: user.ID(100), Status: reminder.StatusCreated},
			},
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			// Setup ---
			unitOfWork := uow.NewFakeUnitOfWork()
			unitOfWork.Reminders().ScheduleResult = testcase.reminders
			scheduler := reminder.NewTestReminderScheduler()
			service := New(
				logging.NewFakeLogger(),
				unitOfWork,
				scheduler,
				func() time.Time { return testcase.now },
			)

			// Exercise ---
			_, err := service.Run(context.Background(), Input{})

			// Verify ---
			s.Nil(err)
			s.ElementsMatch(testcase.reminders, scheduler.Scheduled)
			s.Len(unitOfWork.Reminders().ScheduleWith, 1)
			s.Equal(
				unitOfWork.Reminders().ScheduleWith[0],
				reminder.ScheduleInput{
					ScheduledAt: testcase.now,
					AtBefore:    testcase.now.Add(reminder.DURATION_FOR_SCHEDULING),
				},
			)
			s.True(unitOfWork.Context.WasCommitCalled)
		})
	}
}

func (s *testSuite) TestSchedulingError() {
	s.T().Parallel()

	// Setup ---
	unitOfWork := uow.NewFakeUnitOfWork()
	unitOfWork.Reminders().ScheduleResult = []reminder.Reminder{{ID: reminder.ID(100)}}
	scheduler := reminder.NewTestReminderScheduler()
	scheduler.Error = errors.New("an error occured")
	service := New(
		logging.NewFakeLogger(),
		unitOfWork,
		scheduler,
		func() time.Time { return Now },
	)

	// Exercise ---
	_, err := service.Run(context.Background(), Input{})

	// Verify ---
	assert := s.Require()
	assert.ErrorIs(err, scheduler.Error)
	assert.False(unitOfWork.Context.WasCommitCalled)
	assert.True(unitOfWork.Context.WasRollbackCalled)
}
