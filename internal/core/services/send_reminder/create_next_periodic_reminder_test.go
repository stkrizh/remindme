package sendreminder

import (
	"context"
	"errors"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fixture struct {
	log            *logging.FakeLogger
	unitOfWork     *uow.FakeUnitOfWork
	scheduler      *reminder.TestReminderScheduler
	prepareService *stubPrepareService
}

func newFixture() fixture {
	return fixture{
		log:            logging.NewFakeLogger(),
		unitOfWork:     uow.NewFakeUnitOfWork(),
		scheduler:      reminder.NewTestReminderScheduler(),
		prepareService: newStubPrepareService(),
	}
}

func (f fixture) createService() *createNextPeriodicService {
	return NewCreateNextPeriodicService(
		f.log,
		f.unitOfWork,
		f.scheduler,
		f.prepareService,
	).(*createNextPeriodicService)
}

func TestNewReminderCreated(t *testing.T) {
	cases := []struct {
		id                     string
		every                  reminder.Every
		createdBy              int64
		at                     time.Time
		status                 reminder.Status
		body                   string
		expectedAt             time.Time
		expectedStatus         reminder.Status
		expectedScheduledCount int
	}{
		{
			id:                     "1",
			every:                  reminder.NewEvery(25, reminder.PeriodHour),
			createdBy:              1,
			at:                     time.Date(2023, 1, 1, 15, 30, 20, 1, time.UTC),
			status:                 reminder.StatusSending,
			body:                   "test",
			expectedAt:             time.Date(2023, 1, 2, 16, 30, 20, 1, time.UTC),
			expectedStatus:         reminder.StatusCreated,
			expectedScheduledCount: 0,
		},
		{
			id:                     "2",
			every:                  reminder.EveryHour,
			createdBy:              1,
			at:                     time.Date(2023, 1, 1, 15, 30, 20, 1, time.UTC),
			status:                 reminder.StatusSending,
			body:                   "test",
			expectedAt:             time.Date(2023, 1, 1, 16, 30, 20, 1, time.UTC),
			expectedStatus:         reminder.StatusScheduled,
			expectedScheduledCount: 1,
		},
		{
			id:                     "3",
			every:                  reminder.EveryDay,
			createdBy:              2,
			at:                     time.Date(2023, 1, 1, 15, 30, 20, 1, time.UTC),
			status:                 reminder.StatusSentLimitExceeded,
			body:                   "",
			expectedAt:             time.Date(2023, 1, 2, 15, 30, 20, 1, time.UTC),
			expectedStatus:         reminder.StatusCreated,
			expectedScheduledCount: 0,
		},
		{
			id:                     "4",
			every:                  reminder.EveryMinute,
			createdBy:              3,
			at:                     time.Date(2023, 1, 1, 15, 30, 20, 1, time.UTC),
			status:                 reminder.StatusSending,
			body:                   " ",
			expectedAt:             time.Date(2023, 1, 1, 15, 31, 20, 1, time.UTC),
			expectedStatus:         reminder.StatusScheduled,
			expectedScheduledCount: 1,
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			fixture := newFixture()
			fixture.prepareService.result.Reminder.ID = reminder.ID(321)
			fixture.prepareService.result.Reminder.Every = c.NewOptional(testcase.every, true)
			fixture.prepareService.result.Reminder.CreatedBy = user.ID(testcase.createdBy)
			fixture.prepareService.result.Reminder.At = testcase.at
			fixture.prepareService.result.Reminder.Status = testcase.status
			fixture.prepareService.result.Reminder.Body = testcase.body
			fixture.unitOfWork.Reminders().CreatedID = reminder.ID(123)
			service := fixture.createService()

			_, err := service.Run(context.Background(), Input{})

			assert := require.New(t)
			assert.Nil(err)
			assert.Equal(1, fixture.unitOfWork.Reminders().CreatedCount)
			assert.Equal(c.NewOptional(testcase.every, true), fixture.unitOfWork.Reminders().Created.Every)
			assert.Equal(user.ID(testcase.createdBy), fixture.unitOfWork.Reminders().Created.CreatedBy)
			assert.Equal(testcase.expectedAt, fixture.unitOfWork.Reminders().Created.At)
			assert.Equal(testcase.expectedStatus, fixture.unitOfWork.Reminders().Created.Status)
			assert.Equal(testcase.body, fixture.unitOfWork.Reminders().Created.Body)

			assert.True(fixture.unitOfWork.Context.WasCommitCalled)

			assert.Len(fixture.scheduler.Scheduled, testcase.expectedScheduledCount)
			if testcase.expectedScheduledCount > 0 {
				assert.Equal(reminder.ID(123), fixture.scheduler.Scheduled[0].ID)
			}
		})
	}
}

func TestNewReminderIsNotCreatedIfSentReminderIsNotPeriodic(t *testing.T) {
	fixture := newFixture()
	service := fixture.createService()

	_, err := service.Run(context.Background(), Input{})

	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(0, fixture.unitOfWork.Reminders().CreatedCount)
	assert.Len(fixture.scheduler.Scheduled, 0)
}

func TestNewReminderIsNotCreatedIfPrepareServiceReturnedError(t *testing.T) {
	fixture := newFixture()
	fixture.prepareService.err = errors.New("unexpected error")
	service := fixture.createService()

	_, err := service.Run(context.Background(), Input{})

	assert := require.New(t)
	assert.Error(err)
	assert.Equal(0, fixture.unitOfWork.Reminders().CreatedCount)
	assert.Len(fixture.scheduler.Scheduled, 0)
}

func TestNewReminderCreatedIfSendingLimitExceeded(t *testing.T) {
	fixture := newFixture()
	fixture.prepareService.result.Reminder.Every.IsPresent = true
	fixture.prepareService.result.Reminder.Every.Value = reminder.EveryDay
	fixture.prepareService.err = user.ErrLimitSentReminderCountExceeded
	service := fixture.createService()

	_, err := service.Run(context.Background(), Input{})

	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(1, fixture.unitOfWork.Reminders().CreatedCount)
}

func TestNewReminderCreatedWithChannels(t *testing.T) {
	fixture := newFixture()
	fixture.prepareService.result.Reminder.ID = reminder.ID(123456)
	fixture.prepareService.result.Reminder.Every.IsPresent = true
	fixture.prepareService.result.Reminder.Every.Value = reminder.EveryDay
	fixture.prepareService.result.Reminder.ChannelIDs = []channel.ID{channel.ID(10), channel.ID(20)}
	service := fixture.createService()

	_, err := service.Run(context.Background(), Input{})

	assert := require.New(t)
	assert.Nil(err)
	assert.True(fixture.unitOfWork.ReminderChannels().WasCreateCalled)
	assert.Equal(reminder.ID(0), fixture.unitOfWork.ReminderChannels().CreatedForReminder)
}
