package sendreminder

import (
	"context"
	"errors"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type stubPrepareService struct {
	result Result
	err    error
}

func newStubPrepareService() *stubPrepareService {
	service := &stubPrepareService{}
	service.result.Reminder.Status = reminder.StatusSending
	return service
}

func (s *stubPrepareService) Run(ctx context.Context, input Input) (Result, error) {
	return s.result, s.err
}

func TestReminderSentSuccessfully(t *testing.T) {
	// Setup ---
	log := logging.NewFakeLogger()
	reminderRepo := reminder.NewTestReminderRepository()
	sender := reminder.NewTestReminderSender()
	prepareService := newStubPrepareService()
	service := NewSendService(log, reminderRepo, sender, func() time.Time { return Now }, prepareService)

	// Exercise ---
	result, err := service.Run(context.Background(), Input{ReminderID: REMINDER_ID, At: Now})

	// Verify ---
	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(reminder.StatusSentSuccess, result.Reminder.Status)
	assert.Equal(c.NewOptional(Now, true), result.Reminder.SentAt)
	assert.Len(sender.Sent, 1)
}

func TestReminderNotSentIfStatusIsNotSending(t *testing.T) {
	// Setup ---
	log := logging.NewFakeLogger()
	reminderRepo := reminder.NewTestReminderRepository()
	sender := reminder.NewTestReminderSender()
	prepareService := newStubPrepareService()
	prepareService.result.Reminder.Status = reminder.StatusScheduled
	service := NewSendService(log, reminderRepo, sender, func() time.Time { return Now }, prepareService)

	// Exercise ---
	result, err := service.Run(context.Background(), Input{ReminderID: REMINDER_ID, At: Now})

	// Verify ---
	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(reminder.StatusScheduled, result.Reminder.Status)
	assert.False(result.Reminder.SentAt.IsPresent)
	assert.Len(sender.Sent, 0)
}

func TestReminderNotSentIfInnerServiceReturnsError(t *testing.T) {
	// Setup ---
	log := logging.NewFakeLogger()
	reminderRepo := reminder.NewTestReminderRepository()
	sender := reminder.NewTestReminderSender()
	prepareService := newStubPrepareService()
	prepareService.err = errors.New("test error")
	service := NewSendService(log, reminderRepo, sender, func() time.Time { return Now }, prepareService)

	// Exercise ---
	_, err := service.Run(context.Background(), Input{ReminderID: REMINDER_ID, At: Now})

	// Verify ---
	assert := require.New(t)
	assert.ErrorIs(err, prepareService.err)
	assert.Len(sender.Sent, 0)
}

func TestReminderSendingError(t *testing.T) {
	// Setup ---
	log := logging.NewFakeLogger()
	reminderRepo := reminder.NewTestReminderRepository()
	sender := reminder.NewTestReminderSender()
	sender.SentError = errors.New("test error")
	prepareService := newStubPrepareService()
	service := NewSendService(log, reminderRepo, sender, func() time.Time { return Now }, prepareService)

	// Exercise ---
	result, err := service.Run(context.Background(), Input{ReminderID: REMINDER_ID, At: Now})

	// Verify ---
	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(reminder.StatusSentError, result.Reminder.Status)
	assert.Equal(c.NewOptional(Now, true), result.Reminder.SentAt)
}
