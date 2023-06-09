package sendreminder

import (
	"context"
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

const (
	REMINDER_ID = reminder.ID(123)
	CHANNEL_ID  = channel.ID(321)
)

var (
	Now = time.Date(2020, 6, 6, 15, 30, 0, 0, time.UTC)
)

func TestReminderPreparedSuccessfully(t *testing.T) {
	cases := []struct {
		id         string
		userLimits user.Limits
		sentCount  uint
		reminderAt time.Time
		inputAt    time.Time
	}{
		{
			id:         "1",
			userLimits: user.Limits{},
			sentCount:  0,
			reminderAt: Now,
			inputAt:    Now,
		},
		{
			id:         "2",
			userLimits: user.Limits{MonthlySentReminderCount: c.NewOptional(uint32(100), true)},
			sentCount:  0,
			reminderAt: Now,
			inputAt:    Now,
		},
		{
			id:         "3",
			userLimits: user.Limits{MonthlySentReminderCount: c.NewOptional(uint32(100), true)},
			sentCount:  99,
			reminderAt: Now,
			inputAt:    Now,
		},
		{
			id:         "4",
			userLimits: user.Limits{MonthlySentReminderCount: c.NewOptional(uint32(100), true)},
			sentCount:  50,
			reminderAt: time.Date(2020, 6, 6, 15, 30, 0, 800, time.UTC),
			inputAt:    Now,
		},
		{
			id:         "5",
			userLimits: user.Limits{MonthlySentReminderCount: c.NewOptional(uint32(100), true)},
			sentCount:  50,
			reminderAt: Now,
			inputAt:    time.Date(2020, 6, 6, 15, 30, 0, 800, time.UTC),
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			// Setup ---
			log := logging.NewFakeLogger()
			unitOfWork := uow.NewFakeUnitOfWork()
			unitOfWork.Reminders().GetByIDReminder.ID = REMINDER_ID
			unitOfWork.Reminders().GetByIDReminder.Status = reminder.StatusScheduled
			unitOfWork.Reminders().GetByIDReminder.At = testcase.reminderAt
			unitOfWork.Reminders().GetByIDReminder.ChannelIDs = []channel.ID{CHANNEL_ID}
			unitOfWork.Reminders().CountResult = testcase.sentCount
			unitOfWork.Limits().Limits = testcase.userLimits
			service := NewPrepareService(log, unitOfWork, func() time.Time { return Now })

			// Exercise ---
			result, err := service.Run(context.Background(), Input{ReminderID: REMINDER_ID, At: testcase.inputAt})

			// Verify ---
			assert := require.New(t)
			assert.Nil(err)
			assert.Equal(reminder.StatusSending, result.Reminder.Status)
			assert.True(unitOfWork.Context.WasCommitCalled)
		})
	}
}

func TestReminderNotFound(t *testing.T) {
	// Setup ---
	log := logging.NewFakeLogger()
	unitOfWork := uow.NewFakeUnitOfWork()
	unitOfWork.Reminders().GetByIDError = reminder.ErrReminderDoesNotExist
	service := NewPrepareService(log, unitOfWork, func() time.Time { return Now })

	// Exercise ---
	_, err := service.Run(context.Background(), Input{ReminderID: REMINDER_ID, At: Now})

	// Verify ---
	assert := require.New(t)
	assert.Nil(err)
	assert.False(unitOfWork.Context.WasCommitCalled)
	assert.True(unitOfWork.Context.WasRollbackCalled)
}

func TestReminderStatusIsNotScheduled(t *testing.T) {
	// Setup ---
	log := logging.NewFakeLogger()
	unitOfWork := uow.NewFakeUnitOfWork()
	unitOfWork.Reminders().GetByIDReminder.Status = reminder.StatusCanceled
	service := NewPrepareService(log, unitOfWork, func() time.Time { return Now })

	// Exercise ---
	result, err := service.Run(context.Background(), Input{ReminderID: REMINDER_ID, At: Now})

	// Verify ---
	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(reminder.StatusCanceled, result.Reminder.Status)
	assert.False(unitOfWork.Context.WasCommitCalled)
	assert.True(unitOfWork.Context.WasRollbackCalled)
}

func TestReminderStatusAtTimeChanged(t *testing.T) {
	// Setup ---
	log := logging.NewFakeLogger()
	unitOfWork := uow.NewFakeUnitOfWork()
	unitOfWork.Reminders().GetByIDReminder.Status = reminder.StatusScheduled
	unitOfWork.Reminders().GetByIDReminder.At = Now.Add(6 * time.Hour)
	service := NewPrepareService(log, unitOfWork, func() time.Time { return Now })

	// Exercise ---
	result, err := service.Run(context.Background(), Input{ReminderID: REMINDER_ID, At: Now})

	// Verify ---
	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(reminder.StatusScheduled, result.Reminder.Status)
	assert.False(unitOfWork.Context.WasCommitCalled)
	assert.True(unitOfWork.Context.WasRollbackCalled)
}

func TestUserMonthlySentLimitExceeded(t *testing.T) {
	// Setup ---
	log := logging.NewFakeLogger()
	unitOfWork := uow.NewFakeUnitOfWork()
	unitOfWork.Reminders().GetByIDReminder.Status = reminder.StatusScheduled
	unitOfWork.Reminders().GetByIDReminder.At = Now
	unitOfWork.Limits().Limits.MonthlySentReminderCount = c.NewOptional(uint32(100), true)
	unitOfWork.Reminders().CountResult = 100
	service := NewPrepareService(log, unitOfWork, func() time.Time { return Now })

	// Exercise ---
	result, err := service.Run(context.Background(), Input{ReminderID: REMINDER_ID, At: Now})

	// Verify ---
	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(reminder.StatusSentLimitExceeded, result.Reminder.Status)
	assert.Equal(c.NewOptional(Now, true), result.Reminder.CanceledAt)
	assert.True(unitOfWork.Context.WasCommitCalled)
}
