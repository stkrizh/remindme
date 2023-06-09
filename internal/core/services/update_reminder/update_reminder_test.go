package updatereminder

import (
	"context"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	USER_ID     = 1
	REMINDER_ID = 2
)

var (
	Now        = time.Now().UTC()
	UserLimits = user.Limits{
		ActiveReminderCount:      c.NewOptional(uint32(5), true),
		MonthlySentReminderCount: c.NewOptional(uint32(100), true),
		ReminderEveryPerDayCount: c.NewOptional(float64(1), true),
	}
	ChannelIDs = []channel.ID{channel.ID(1), channel.ID(10)}
)

type testSuite struct {
	suite.Suite
	logger     *logging.FakeLogger
	unitOfWork *uow.FakeUnitOfWork
	scheduler  *reminder.TestReminderScheduler
	service    services.Service[Input, Result]
	input      Input
}

func (suite *testSuite) SetupTest() {
	suite.logger = logging.NewFakeLogger()
	suite.unitOfWork = uow.NewFakeUnitOfWork()
	suite.unitOfWork.Limits().Limits = UserLimits
	suite.unitOfWork.Reminders().GetByIDReminder.CreatedBy = USER_ID
	suite.unitOfWork.Reminders().GetByIDReminder.Status = reminder.StatusCreated
	suite.unitOfWork.Reminders().GetByIDReminder.ChannelIDs = ChannelIDs
	suite.scheduler = reminder.NewTestReminderScheduler()
	suite.service = New(
		suite.logger,
		suite.unitOfWork,
		suite.scheduler,
		func() time.Time { return Now },
	)
	suite.input = Input{
		UserID:     USER_ID,
		ReminderID: REMINDER_ID,
	}
}

func (suite *testSuite) TearDownTest() {
	suite.scheduler.Scheduled = make([]reminder.Reminder, 0)
}

func TestUpdateReminderService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestValidateAt() {
	cases := []struct {
		id            string
		at            time.Time
		expectedError error
	}{
		{
			id:            "1",
			at:            Now.Add(reminder.MIN_DURATION_FROM_NOW),
			expectedError: nil,
		},
		{
			id:            "2",
			at:            Now.Add(reminder.MAX_DURATION_FROM_NOW),
			expectedError: nil,
		},
		{
			id:            "3",
			at:            Now.Add(reminder.MIN_DURATION_FROM_NOW).Add(-time.Second),
			expectedError: reminder.ErrReminderTooEarly,
		},
		{
			id:            "4",
			at:            Now.Add(reminder.MAX_DURATION_FROM_NOW).Add(time.Second),
			expectedError: reminder.ErrReminderTooLate,
		},
		{
			id:            "5",
			at:            time.Date(2020, 10, 15, 15, 40, 2, 0, mustLoadLocation("America/New_York")),
			expectedError: reminder.ErrReminderAtTimeIsNotUTC,
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			err := validateAt(testcase.at, Now)
			s.ErrorIs(err, testcase.expectedError)
		})
	}
}

func (s *testSuite) TestValidateEvery() {
	cases := []struct {
		id              string
		every           c.Optional[reminder.Every]
		userPerDayLimit c.Optional[float64]
		expectedError   error
	}{
		{
			id:              "1",
			every:           c.NewOptional(reminder.EveryDay, true),
			userPerDayLimit: c.NewOptional(1.0, true),
			expectedError:   nil,
		},
		{
			id:              "2",
			every:           c.NewOptional(reminder.EveryHour, true),
			userPerDayLimit: c.NewOptional(24.0, true),
			expectedError:   nil,
		},
		{
			id:              "3",
			every:           c.NewOptional(reminder.EveryMinute, true),
			userPerDayLimit: c.NewOptional(0.0, false),
			expectedError:   nil,
		},
		{
			id:              "4",
			every:           c.NewOptional(reminder.EveryHour, true),
			userPerDayLimit: c.NewOptional(1.0, true),
			expectedError:   user.ErrLimitReminderEveryPerDayCountExceeded,
		},
		{
			id:              "5",
			every:           c.NewOptional(reminder.EveryHour, true),
			userPerDayLimit: c.NewOptional(23.0, true),
			expectedError:   user.ErrLimitReminderEveryPerDayCountExceeded,
		},
		{
			id:              "6",
			every:           c.NewOptional(reminder.Every{}, false),
			userPerDayLimit: c.NewOptional(1.0, true),
			expectedError:   nil,
		},
		{
			id:              "7",
			every:           c.NewOptional(reminder.NewEvery(2, reminder.PeriodYear), true),
			userPerDayLimit: c.NewOptional(1.0, true),
			expectedError:   reminder.ErrInvalidEvery,
		},
		{
			id:              "8",
			every:           c.NewOptional(reminder.NewEvery(12, reminder.PeriodMonth), true),
			userPerDayLimit: c.NewOptional(1.0, true),
			expectedError:   nil,
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			repo := user.NewFakeLimitsRepository()
			repo.Limits = user.Limits{ReminderEveryPerDayCount: testcase.userPerDayLimit}

			err := validateEvery(context.Background(), s.logger, USER_ID, testcase.every, repo)
			s.ErrorIs(err, testcase.expectedError)
		})
	}
}

func (s *testSuite) TestUpdateSuccess() {
	cases := []struct {
		id                string
		statusBefore      reminder.Status
		everyBefore       c.Optional[reminder.Every]
		bodyBefore        string
		scheduledAtBefore c.Optional[time.Time]
		doAtUpdate        bool
		at                time.Time
		doEveryUpdate     bool
		every             c.Optional[reminder.Every]
		doBodyUpdate      bool
		body              string
		scheduled         int
		statusAfter       reminder.Status
		scheduledAtAfter  c.Optional[time.Time]
	}{
		{
			id:           "1",
			statusBefore: reminder.StatusCreated,
			statusAfter:  reminder.StatusCreated,
		},
		{
			id:           "2",
			statusBefore: reminder.StatusCreated,
			statusAfter:  reminder.StatusCreated,
		},
		{
			id:           "3",
			statusBefore: reminder.StatusCreated,
			doAtUpdate:   true,
			at:           Now.Add(reminder.DURATION_FOR_SCHEDULING + time.Hour),
			statusAfter:  reminder.StatusCreated,
		},
		{
			id:               "4",
			statusBefore:     reminder.StatusCreated,
			doAtUpdate:       true,
			at:               Now.Add(reminder.DURATION_FOR_SCHEDULING - time.Second),
			scheduled:        1,
			statusAfter:      reminder.StatusScheduled,
			scheduledAtAfter: c.NewOptional(Now, true),
		},
		{
			id:                "5",
			statusBefore:      reminder.StatusScheduled,
			scheduledAtBefore: c.NewOptional(Now.Add(-1*time.Hour), true),
			doAtUpdate:        true,
			at:                Now.Add(reminder.DURATION_FOR_SCHEDULING - time.Hour),
			scheduled:         1,
			statusAfter:       reminder.StatusScheduled,
			scheduledAtAfter:  c.NewOptional(Now, true),
		},
		{
			id:                "6",
			statusBefore:      reminder.StatusScheduled,
			scheduledAtBefore: c.NewOptional(Now.Add(-1*time.Hour), true),
			doAtUpdate:        true,
			at:                Now.Add(reminder.DURATION_FOR_SCHEDULING + 120*time.Hour),
			scheduled:         0,
			statusAfter:       reminder.StatusCreated,
		},
		{
			id:            "7",
			statusBefore:  reminder.StatusCreated,
			everyBefore:   c.NewOptional(reminder.EveryWeek, true),
			doEveryUpdate: true,
			every:         c.NewOptional(reminder.Every{}, false),
			scheduled:     0,
			statusAfter:   reminder.StatusCreated,
		},
		{
			id:            "8",
			statusBefore:  reminder.StatusCreated,
			everyBefore:   c.NewOptional(reminder.EveryWeek, true),
			doEveryUpdate: true,
			every:         c.NewOptional(reminder.NewEvery(25, reminder.PeriodHour), true),
			scheduled:     0,
			statusAfter:   reminder.StatusCreated,
		},
		{
			id:            "9",
			statusBefore:  reminder.StatusCreated,
			everyBefore:   c.NewOptional(reminder.Every{}, false),
			doEveryUpdate: true,
			every:         c.NewOptional(reminder.NewEvery(25, reminder.PeriodHour), true),
			scheduled:     0,
			statusAfter:   reminder.StatusCreated,
		},
		{
			id:           "10",
			statusBefore: reminder.StatusCreated,
			bodyBefore:   "",
			doBodyUpdate: true,
			body:         "",
			scheduled:    0,
			statusAfter:  reminder.StatusCreated,
		},
		{
			id:           "11",
			statusBefore: reminder.StatusCreated,
			bodyBefore:   "",
			doBodyUpdate: true,
			body:         "aaa",
			scheduled:    0,
			statusAfter:  reminder.StatusCreated,
		},
		{
			id:           "12",
			statusBefore: reminder.StatusCreated,
			bodyBefore:   "aaa",
			doBodyUpdate: true,
			body:         "",
			scheduled:    0,
			statusAfter:  reminder.StatusCreated,
		},
		{
			id:            "13",
			statusBefore:  reminder.StatusCreated,
			everyBefore:   c.NewOptional(reminder.Every{}, false),
			bodyBefore:    "",
			doAtUpdate:    true,
			at:            Now.Add(240 * time.Hour),
			doEveryUpdate: true,
			every:         c.NewOptional(reminder.EveryDay, true),
			doBodyUpdate:  true,
			body:          "aaa",
			scheduled:     0,
			statusAfter:   reminder.StatusCreated,
		},
		{
			id:           "14",
			statusBefore: reminder.StatusScheduled,
			doAtUpdate:   true,
			at:           Now,
			scheduled:    0,
			statusAfter:  reminder.StatusScheduled,
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			s.SetupTest()
			input := s.input
			input.DoAtUpdate = testcase.doAtUpdate
			input.At = testcase.at
			input.DoEveryUpdate = testcase.doEveryUpdate
			input.Every = testcase.every
			input.DoBodyUpdate = testcase.doBodyUpdate
			input.Body = testcase.body
			s.unitOfWork.Reminders().GetByIDReminder.Status = testcase.statusBefore
			s.unitOfWork.Reminders().GetByIDReminder.At = Now
			s.unitOfWork.Reminders().ReminderBeforeUpdate.Status = testcase.statusBefore
			s.unitOfWork.Reminders().ReminderBeforeUpdate.ScheduledAt = testcase.scheduledAtBefore
			s.unitOfWork.Reminders().ReminderBeforeUpdate.At = Now
			s.unitOfWork.Reminders().ReminderBeforeUpdate.Every = testcase.everyBefore
			s.unitOfWork.Reminders().ReminderBeforeUpdate.Body = testcase.bodyBefore

			result, err := s.service.Run(context.Background(), input)

			s.Nil(err)
			s.Equal(reminder.ID(REMINDER_ID), result.Reminder.ID)
			s.Equal(ChannelIDs, result.Reminder.ChannelIDs)
			s.True(s.unitOfWork.Context.WasCommitCalled)
			s.Equal(testcase.statusAfter, result.Reminder.Status)
			s.Equal(testcase.scheduledAtAfter, result.Reminder.ScheduledAt)
			s.Len(s.scheduler.Scheduled, testcase.scheduled)

			if testcase.doAtUpdate {
				s.Equal(testcase.at, result.Reminder.At)
			} else {
				s.Equal(Now, result.Reminder.At)
			}
			if testcase.doEveryUpdate {
				s.Equal(testcase.every, result.Reminder.Every)
			} else {
				s.Equal(testcase.everyBefore, result.Reminder.Every)
			}
			if testcase.doBodyUpdate {
				s.Equal(testcase.body, result.Reminder.Body)
			} else {
				s.Equal(testcase.bodyBefore, result.Reminder.Body)
			}
		})
	}
}

func (s *testSuite) TestItsNotPossibleToUpdateOtherUserReminder() {
	s.unitOfWork.Reminders().GetByIDReminder.CreatedBy = user.ID(111222333)

	_, err := s.service.Run(context.Background(), s.input)
	s.ErrorIs(err, reminder.ErrReminderPermission)
	s.False(s.unitOfWork.Context.WasCommitCalled)
	s.Len(s.scheduler.Scheduled, 0)
}

func (s *testSuite) TestItsNotPossibleToNotActiveReminder() {
	statuses := []reminder.Status{
		reminder.StatusSending,
		reminder.StatusSentSuccess,
		reminder.StatusSentError,
		reminder.StatusSentLimitExceeded,
		reminder.StatusCanceled,
	}
	for _, status := range statuses {
		s.unitOfWork.Reminders().GetByIDReminder.Status = status

		_, err := s.service.Run(context.Background(), s.input)
		assert := s.Require()
		assert.ErrorIs(err, reminder.ErrReminderNotActive, status)
		assert.False(s.unitOfWork.Context.WasCommitCalled, status)
		assert.Len(s.scheduler.Scheduled, 0, status)
	}
}

func (s *testSuite) TestItsNotPossibleToUpdateIfNewAtIsInvalid() {
	s.input.DoAtUpdate = true
	s.input.At = Now.Add(-1 * time.Hour)

	_, err := s.service.Run(context.Background(), s.input)
	s.ErrorIs(err, reminder.ErrReminderTooEarly)
	s.False(s.unitOfWork.Context.WasCommitCalled)
	s.Len(s.scheduler.Scheduled, 0)
}

func (s *testSuite) TestItsNotPossibleToUpdateIfNewEveryIsInvalid() {
	s.input.DoEveryUpdate = true
	s.input.Every = c.NewOptional(reminder.NewEvery(10, reminder.PeriodYear), true)

	_, err := s.service.Run(context.Background(), s.input)
	s.ErrorIs(err, reminder.ErrInvalidEvery)
	s.False(s.unitOfWork.Context.WasCommitCalled)
	s.Len(s.scheduler.Scheduled, 0)
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

func TestDoAtUpdate(t *testing.T) {
	cases := []struct {
		id       string
		input    Input
		now      time.Time
		doUpdate bool
	}{
		{
			id:       "1",
			input:    Input{DoAtUpdate: true, At: time.Date(2020, 1, 1, 1, 1, 1, 123, time.UTC)},
			now:      time.Date(2020, 1, 1, 1, 1, 1, 123, time.UTC),
			doUpdate: false,
		},
		{
			id:       "2",
			input:    Input{DoAtUpdate: true, At: time.Date(2020, 1, 1, 1, 1, 1, 123, time.UTC)},
			now:      time.Date(2020, 1, 1, 1, 1, 1, 123456, time.UTC),
			doUpdate: false,
		},
		{
			id:       "3",
			input:    Input{DoAtUpdate: false, At: time.Date(2025, 1, 1, 1, 1, 1, 123, time.UTC)},
			now:      time.Date(2020, 1, 1, 1, 1, 1, 123456, time.UTC),
			doUpdate: false,
		},
		{
			id:       "4",
			input:    Input{DoAtUpdate: true, At: time.Date(2025, 1, 1, 1, 1, 2, 123, time.UTC)},
			now:      time.Date(2020, 1, 1, 1, 1, 1, 123456, time.UTC),
			doUpdate: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.id, func(t *testing.T) {
			assert := require.New(t)
			assert.Equal(tc.doUpdate, doAtUpdate(tc.input, tc.now))
		})
	}
}
