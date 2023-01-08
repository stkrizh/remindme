package cancelreminder

import (
	"context"
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
	USER_ID     = user.ID(42)
	REMINDER_ID = reminder.ID(77)
)

var (
	Now time.Time = time.Now().UTC()
)

type testSuite struct {
	suite.Suite
	logger     *logging.FakeLogger
	unitOfWork *uow.FakeUnitOfWork
	service    services.Service[Input, Result]
}

func (suite *testSuite) SetupTest() {
	suite.logger = logging.NewFakeLogger()
	suite.unitOfWork = uow.NewFakeUnitOfWork()
	suite.service = New(
		suite.logger,
		suite.unitOfWork,
		func() time.Time { return Now },
	)
}

func (suite *testSuite) TearDownTest() {}

func TestCancelReminderService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestCancelSuccess() {
	cases := []struct {
		id         string
		status     reminder.Status
		reminderID reminder.ID
		userID     user.ID
	}{
		{id: "1", status: reminder.StatusCreated, reminderID: reminder.ID(100), userID: user.ID(1)},
		{id: "2", status: reminder.StatusScheduled, reminderID: reminder.ID(200), userID: user.ID(2)},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			s.SetupTest()
			s.unitOfWork.Reminders().GetByIDReminder.ID = testcase.reminderID
			s.unitOfWork.Reminders().GetByIDReminder.CreatedBy = testcase.userID
			s.unitOfWork.Reminders().GetByIDReminder.Status = testcase.status
			s.unitOfWork.Reminders().GetByIDReminder.CanceledAt.IsPresent = false

			result, err := s.service.Run(
				context.Background(),
				Input{UserID: testcase.userID, ReminderID: testcase.reminderID},
			)

			assert := s.Require()
			assert.Nil(err)
			assert.Equal(testcase.reminderID, result.Reminder.ID)
			assert.Equal(reminder.StatusCanceled, result.Reminder.Status)
			assert.Equal(c.NewOptional(Now, true), result.Reminder.CanceledAt)

			assert.True(s.unitOfWork.Context.WasCommitCalled)
		})
	}
}

func (s *testSuite) TestCancelError() {
	cases := []struct {
		id            string
		getByIDError  error
		status        reminder.Status
		userID        user.ID
		expectedError error
	}{
		{
			id:            "1",
			status:        reminder.StatusCreated,
			userID:        user.ID(1),
			expectedError: reminder.ErrReminderPermission,
		},
		{
			id:            "2",
			status:        reminder.StatusSendSuccess,
			userID:        USER_ID,
			expectedError: reminder.ErrReminderNotActive,
		},
		{
			id:            "3",
			status:        reminder.StatusSendError,
			userID:        USER_ID,
			expectedError: reminder.ErrReminderNotActive,
		},
		{
			id:            "4",
			status:        reminder.StatusSendLimitExceeded,
			userID:        USER_ID,
			expectedError: reminder.ErrReminderNotActive,
		},
		{
			id:            "5",
			status:        reminder.StatusCanceled,
			userID:        USER_ID,
			expectedError: reminder.ErrReminderNotActive,
		},
		{
			id:            "6",
			getByIDError:  reminder.ErrReminderDoesNotExist,
			status:        reminder.StatusCreated,
			userID:        USER_ID,
			expectedError: reminder.ErrReminderDoesNotExist,
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			s.SetupTest()
			s.unitOfWork.Reminders().GetByIDError = testcase.getByIDError
			s.unitOfWork.Reminders().GetByIDReminder.CreatedBy = USER_ID
			s.unitOfWork.Reminders().GetByIDReminder.Status = testcase.status

			_, err := s.service.Run(
				context.Background(),
				Input{UserID: testcase.userID, ReminderID: REMINDER_ID},
			)

			assert := s.Require()
			assert.ErrorIs(err, testcase.expectedError)

			assert.False(s.unitOfWork.Context.WasCommitCalled)
		})
	}
}
