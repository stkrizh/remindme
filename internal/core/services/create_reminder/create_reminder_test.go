package createreminder

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

	"github.com/stretchr/testify/suite"
)

const (
	USER_ID      = user.ID(42)
	CHANNEL_ID_1 = channel.ID(1)
	CHANNEL_ID_2 = channel.ID(2)
)

var (
	Now        time.Time = time.Now().UTC()
	UserLimits           = user.Limits{
		ActiveReminderCount:      c.NewOptional(uint32(5), true),
		MonthlySentReminderCount: c.NewOptional(uint32(100), true),
		ReminderEveryPerDayCount: c.NewOptional(float64(1), true),
	}
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
	suite.unitOfWork.Channels().ReadChannels = []channel.Channel{
		{ID: CHANNEL_ID_1, CreatedBy: user.ID(USER_ID), CreatedAt: Now, VerifiedAt: c.NewOptional(Now, true)},
		{ID: CHANNEL_ID_2, CreatedBy: user.ID(USER_ID), CreatedAt: Now, VerifiedAt: c.NewOptional(Now, true)},
	}
	suite.scheduler = reminder.NewTestReminderScheduler()
	suite.service = New(
		suite.logger,
		suite.unitOfWork,
		suite.scheduler,
		func() time.Time { return Now },
	)
	suite.input = Input{
		UserID:     USER_ID,
		ChannelIDs: reminder.NewChannelIDs(CHANNEL_ID_1, CHANNEL_ID_2),
	}
}

func (suite *testSuite) TearDownTest() {
	suite.scheduler.Scheduled = make([]reminder.Reminder, 0)
}

func TestCreateReminderService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestCreateSuccessWithStatusCreated() {
	cases := []struct {
		id                            string
		now                           time.Time
		at                            time.Time
		every                         c.Optional[reminder.Every]
		limitActiveReminderCount      c.Optional[uint32]
		limitMonthlySentReminderCount c.Optional[uint32]
		limitReminderEveryPerDayCount c.Optional[float64]
		actualReminderCount           uint
	}{
		{
			id:  "1",
			now: time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:  time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
		},
		{
			id:    "2",
			now:   time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:    time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			every: c.NewOptional(reminder.EveryHour, true),
		},
		{
			id:    "3",
			now:   time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:    time.Date(2002, 1, 1, 1, 1, 1, 0, time.UTC),
			every: c.NewOptional(reminder.EveryHour, true),
		},
		{
			id:                       "4",
			now:                      time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                       time.Date(2000, 2, 1, 1, 1, 1, 0, time.UTC),
			limitActiveReminderCount: c.NewOptional(uint32(1), true),
		},
		{
			id:                       "5",
			now:                      time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                       time.Date(2000, 2, 1, 1, 1, 1, 0, time.UTC),
			limitActiveReminderCount: c.NewOptional(uint32(2), true),
			actualReminderCount:      1,
		},
		{
			id:                            "6",
			now:                           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			limitActiveReminderCount:      c.NewOptional(uint32(2), true),
			limitMonthlySentReminderCount: c.NewOptional(uint32(100), true),
			actualReminderCount:           1,
		},
		{
			id:                            "7",
			now:                           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			limitMonthlySentReminderCount: c.NewOptional(uint32(100), true),
			actualReminderCount:           99,
		},
		{
			id:                            "8",
			now:                           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                            time.Date(2000, 2, 1, 1, 1, 1, 0, time.UTC),
			limitMonthlySentReminderCount: c.NewOptional(uint32(100), true),
			actualReminderCount:           100,
		},
		{
			id:    "9",
			now:   time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:    time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			every: c.NewOptional(reminder.EveryMinute, true),
		},
		{
			id:                            "10",
			now:                           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			every:                         c.NewOptional(reminder.EveryMinute, true),
			limitReminderEveryPerDayCount: c.NewOptional(24.0*60, true),
		},
		{
			id:                            "11",
			now:                           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			every:                         c.NewOptional(reminder.EveryHour, true),
			limitReminderEveryPerDayCount: c.NewOptional(24.0, true),
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			s.unitOfWork.Limits().Limits = user.Limits{
				ActiveReminderCount:      testcase.limitActiveReminderCount,
				MonthlySentReminderCount: testcase.limitMonthlySentReminderCount,
				ReminderEveryPerDayCount: testcase.limitReminderEveryPerDayCount,
			}
			s.unitOfWork.Reminders().CountResult = testcase.actualReminderCount

			input := s.input
			input.At = testcase.at
			input.Every = testcase.every

			service := New(s.logger, s.unitOfWork, s.scheduler, func() time.Time { return testcase.now })
			result, err := service.Run(context.Background(), input)

			assert := s.Require()
			assert.Nil(err)
			assert.Equal(USER_ID, result.Reminder.CreatedBy)
			assert.Equal(testcase.now, result.Reminder.CreatedAt)
			assert.Equal(testcase.every, result.Reminder.Every)
			assert.Equal(reminder.StatusCreated, result.Reminder.Status)
			assert.ElementsMatch([]channel.ID{CHANNEL_ID_1, CHANNEL_ID_2}, result.Reminder.ChannelIDs)

			assert.True(s.unitOfWork.Context.WasCommitCalled)

			assert.Equal(0, len(s.scheduler.Scheduled))
		})
	}
}

func (s *testSuite) TestCreateSuccessWithStatusScheduled() {
	cases := []struct {
		id  string
		now time.Time
		at  time.Time
	}{
		{
			id:  "1",
			now: time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:  time.Date(2000, 1, 1, 1, 31, 1, 0, time.UTC),
		},
		{
			id:  "2",
			now: time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:  time.Date(2000, 1, 2, 1, 1, 0, 0, time.UTC),
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			defer s.TearDownTest()

			input := s.input
			input.At = testcase.at

			service := New(s.logger, s.unitOfWork, s.scheduler, func() time.Time { return testcase.now })
			result, err := service.Run(context.Background(), input)

			assert := s.Require()
			assert.Nil(err)
			assert.Equal(USER_ID, result.Reminder.CreatedBy)
			assert.Equal(testcase.now, result.Reminder.CreatedAt)
			assert.Equal(reminder.StatusScheduled, result.Reminder.Status)
			assert.Equal(c.NewOptional(testcase.now, true), result.Reminder.ScheduledAt)
			assert.Equal([]channel.ID{CHANNEL_ID_1, CHANNEL_ID_2}, result.Reminder.ChannelIDs)

			assert.True(s.unitOfWork.Context.WasCommitCalled)

			assert.Equal(1, len(s.scheduler.Scheduled))
		})
	}
}

func (s *testSuite) TestCreateError() {
	cases := []struct {
		id                            string
		now                           time.Time
		at                            time.Time
		every                         c.Optional[reminder.Every]
		channelIDs                    []channel.ID
		limitActiveReminderCount      c.Optional[uint32]
		limitMonthlySentReminderCount c.Optional[uint32]
		limitReminderEveryPerDayCount c.Optional[float64]
		actualReminderCount           uint
		expectedError                 error
		wasRollbackCalled             bool
	}{
		{
			id:            "1",
			now:           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:            time.Date(2000, 1, 1, 1, 1, 0, 0, time.UTC),
			expectedError: reminder.ErrReminderTooEarly,
		},
		{
			id:            "2",
			now:           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:            time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			expectedError: reminder.ErrReminderTooEarly,
		},
		{
			id:            "3",
			now:           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:            time.Date(1999, 1, 1, 1, 1, 1, 0, time.UTC),
			expectedError: reminder.ErrReminderTooEarly,
		},
		{
			id:            "4",
			now:           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:            time.Date(2003, 1, 1, 1, 1, 1, 0, time.UTC),
			expectedError: reminder.ErrReminderTooLate,
		},
		{
			id:            "5",
			now:           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			every:         c.NewOptional(reminder.NewEvery(36, reminder.PeriodMonth), true),
			expectedError: reminder.ErrInvalidEvery,
		},
		{
			id:            "6",
			now:           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			expectedError: reminder.ErrReminderChannelsNotSet,
		},
		{
			id:            "7",
			now:           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			expectedError: reminder.ErrReminderTooManyChannels,
			channelIDs: []channel.ID{
				channel.ID(1),
				channel.ID(2),
				channel.ID(3),
				channel.ID(4),
				channel.ID(5),
				channel.ID(6),
			},
		},
		{
			id:                       "8",
			now:                      time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                       time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			channelIDs:               []channel.ID{CHANNEL_ID_1, CHANNEL_ID_2},
			limitActiveReminderCount: c.NewOptional(uint32(5), true),
			actualReminderCount:      5,
			expectedError:            user.ErrLimitActiveReminderCountExceeded,
			wasRollbackCalled:        true,
		},
		{
			id:                       "9",
			now:                      time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                       time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			channelIDs:               []channel.ID{CHANNEL_ID_1, CHANNEL_ID_2},
			limitActiveReminderCount: c.NewOptional(uint32(2), true),
			actualReminderCount:      10,
			expectedError:            user.ErrLimitActiveReminderCountExceeded,
			wasRollbackCalled:        true,
		},
		{
			id:                            "10",
			now:                           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			channelIDs:                    []channel.ID{CHANNEL_ID_1, CHANNEL_ID_2},
			limitActiveReminderCount:      c.NewOptional(uint32(1000), true),
			limitMonthlySentReminderCount: c.NewOptional(uint32(100), true),
			actualReminderCount:           100,
			expectedError:                 user.ErrLimitSentReminderCountExceeded,
			wasRollbackCalled:             true,
		},
		{
			id:                            "11",
			now:                           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			every:                         c.NewOptional(reminder.EveryMinute, true),
			channelIDs:                    []channel.ID{CHANNEL_ID_1, CHANNEL_ID_2},
			limitActiveReminderCount:      c.NewOptional(uint32(10), true),
			limitMonthlySentReminderCount: c.NewOptional(uint32(100), true),
			limitReminderEveryPerDayCount: c.NewOptional(1.0, true),
			actualReminderCount:           1,
			expectedError:                 user.ErrLimitReminderEveryPerDayCountExceeded,
			wasRollbackCalled:             true,
		},
		{
			id:                            "12",
			now:                           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			every:                         c.NewOptional(reminder.NewEvery(59, reminder.PeriodMinute), true),
			channelIDs:                    []channel.ID{CHANNEL_ID_1, CHANNEL_ID_2},
			limitActiveReminderCount:      c.NewOptional(uint32(10), true),
			limitMonthlySentReminderCount: c.NewOptional(uint32(100), true),
			limitReminderEveryPerDayCount: c.NewOptional(24.0, true),
			actualReminderCount:           1,
			expectedError:                 user.ErrLimitReminderEveryPerDayCountExceeded,
			wasRollbackCalled:             true,
		},
		{
			id:                            "13",
			now:                           time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                            time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			channelIDs:                    []channel.ID{CHANNEL_ID_1, CHANNEL_ID_2, channel.ID(111222333)},
			limitActiveReminderCount:      c.NewOptional(uint32(10), true),
			limitMonthlySentReminderCount: c.NewOptional(uint32(100), true),
			limitReminderEveryPerDayCount: c.NewOptional(1.0, true),
			actualReminderCount:           1,
			expectedError:                 reminder.ErrReminderChannelsNotValid,
			wasRollbackCalled:             true,
		},
		{
			id:                "14",
			now:               time.Date(2000, 1, 1, 1, 1, 1, 0, time.UTC),
			at:                time.Date(2000, 1, 2, 1, 1, 1, 0, time.UTC),
			every:             c.NewOptional(reminder.EveryDay, true),
			channelIDs:        nil,
			expectedError:     reminder.ErrReminderChannelsNotSet,
			wasRollbackCalled: true,
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			s.unitOfWork.Limits().Limits = user.Limits{
				ActiveReminderCount:      testcase.limitActiveReminderCount,
				MonthlySentReminderCount: testcase.limitMonthlySentReminderCount,
				ReminderEveryPerDayCount: testcase.limitReminderEveryPerDayCount,
			}
			s.unitOfWork.Reminders().CountResult = testcase.actualReminderCount

			input := s.input
			input.At = testcase.at
			input.Every = testcase.every
			input.ChannelIDs = reminder.NewChannelIDs(testcase.channelIDs...)

			service := New(s.logger, s.unitOfWork, s.scheduler, func() time.Time { return testcase.now })
			_, err := service.Run(context.Background(), input)

			assert := s.Require()
			assert.ErrorIs(err, testcase.expectedError)

			assert.False(s.unitOfWork.Context.WasCommitCalled)
			assert.Equal(testcase.wasRollbackCalled, s.unitOfWork.Context.WasRollbackCalled)

			assert.Equal(0, len(s.scheduler.Scheduled))
		})
	}
}

func (s *testSuite) TestCreateErrorIfChannelsNotVerified() {
	s.unitOfWork.Limits().Limits = user.Limits{}
	s.unitOfWork.Channels().ReadChannels = []channel.Channel{
		{ID: CHANNEL_ID_1, CreatedBy: user.ID(USER_ID), CreatedAt: Now, VerifiedAt: c.NewOptional(Now, true)},
		{ID: CHANNEL_ID_2, CreatedBy: user.ID(USER_ID), CreatedAt: Now},
	}

	input := s.input
	input.At = Now.Add(time.Hour)
	input.ChannelIDs = reminder.NewChannelIDs(CHANNEL_ID_1, CHANNEL_ID_2)
	service := New(s.logger, s.unitOfWork, s.scheduler, func() time.Time { return Now })
	_, err := service.Run(context.Background(), input)

	assert := s.Require()
	assert.ErrorIs(err, reminder.ErrReminderChannelsNotVerified)
}

func (s *testSuite) TestCreateErrorIfAtTimeIsNotUTC() {
	at, err := time.Parse(time.RFC3339, "2020-02-29T01:02:03+04:00")
	s.Nil(err)

	input := s.input
	input.At = at
	input.ChannelIDs = reminder.NewChannelIDs(CHANNEL_ID_1, CHANNEL_ID_2)
	service := New(s.logger, s.unitOfWork, s.scheduler, func() time.Time { return Now })
	_, err = service.Run(context.Background(), input)

	assert := s.Require()
	assert.ErrorIs(err, reminder.ErrReminderAtTimeIsNotUTC)
}
