package updatereminderchannels

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
	USER_ID     = 1
	REMINDER_ID = 2
)

var (
	Now        = time.Now().UTC()
	ChannelIDs = []channel.ID{channel.ID(1), channel.ID(10)}
)

type testSuite struct {
	suite.Suite
	logger     *logging.FakeLogger
	unitOfWork *uow.FakeUnitOfWork
	service    services.Service[Input, Result]
	input      Input
}

func (suite *testSuite) SetupTest() {
	suite.logger = logging.NewFakeLogger()
	suite.unitOfWork = uow.NewFakeUnitOfWork()
	suite.unitOfWork.Reminders().GetByIDReminder = reminder.ReminderWithChannels{
		Reminder: reminder.Reminder{
			Status:    reminder.StatusCreated,
			CreatedBy: USER_ID,
		},
		ChannelIDs: []channel.ID{channel.ID(1)},
	}
	suite.unitOfWork.Channels().ReadChannels = []channel.Channel{
		{ID: channel.ID(111), CreatedBy: USER_ID, VerifiedAt: c.NewOptional(Now, true)},
		{ID: channel.ID(222), CreatedBy: USER_ID, VerifiedAt: c.NewOptional(Now, true)},
	}
	suite.service = New(
		suite.logger,
		suite.unitOfWork,
		func() time.Time { return Now },
	)
	suite.input = Input{
		UserID:     USER_ID,
		ReminderID: REMINDER_ID,
		ChannelIDs: reminder.NewChannelIDs(channel.ID(111), channel.ID(222)),
	}
}

func (suite *testSuite) TearDownTest() {}

func TestUpdateReminderChannelsService(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestValidateInput() {
	cases := []struct {
		id            string
		channelIDs    []channel.ID
		expectedError error
	}{
		{
			id:         "1",
			channelIDs: []channel.ID{channel.ID(1)},
		},
		{
			id:         "2",
			channelIDs: []channel.ID{channel.ID(1), channel.ID(100)},
		},
		{
			id:            "3",
			channelIDs:    []channel.ID{},
			expectedError: reminder.ErrReminderChannelsNotSet,
		},
		{
			id: "4",
			channelIDs: []channel.ID{
				channel.ID(1),
				channel.ID(2),
				channel.ID(3),
				channel.ID(4),
				channel.ID(5),
				channel.ID(6),
			},
			expectedError: reminder.ErrReminderTooManyChannels,
		},
	}

	for _, testcacases := range cases {
		s.Run(testcacases.id, func() {
			input := s.input
			input.ChannelIDs = reminder.NewChannelIDs(testcacases.channelIDs...)

			err := input.Validate()
			s.ErrorIs(err, testcacases.expectedError)
		})
	}
}

func (s *testSuite) TestSuccess() {
	result, err := s.service.Run(context.Background(), s.input)

	s.Nil(err)
	s.ElementsMatch([]channel.ID{channel.ID(111), channel.ID(222)}, result.ChannelIDs)
	s.True(s.unitOfWork.Context.WasCommitCalled)
	s.Equal(s.input.ReminderID, s.unitOfWork.ReminderChannels().CreatedForReminder)
	s.Equal(s.input.ReminderID, s.unitOfWork.ReminderChannels().DeletedByReminderID)
	s.True(s.unitOfWork.ReminderChannels().WasCreateCalled)
	s.True(s.unitOfWork.ReminderChannels().WasDeleteCalled)
}

func (s *testSuite) TestErrorReminderDoesNotExist() {
	s.unitOfWork.Reminders().GetByIDError = reminder.ErrReminderDoesNotExist

	_, err := s.service.Run(context.Background(), s.input)

	s.ErrorIs(err, reminder.ErrReminderDoesNotExist)
	s.True(s.unitOfWork.Context.WasRollbackCalled)
	s.False(s.unitOfWork.Context.WasCommitCalled)
	s.False(s.unitOfWork.ReminderChannels().WasCreateCalled)
	s.False(s.unitOfWork.ReminderChannels().WasDeleteCalled)
}

func (s *testSuite) TestErrorReminderBelongsToOtherUser() {
	s.unitOfWork.Reminders().GetByIDReminder.CreatedBy = user.ID(111222333)

	_, err := s.service.Run(context.Background(), s.input)

	s.ErrorIs(err, reminder.ErrReminderPermission)
	s.True(s.unitOfWork.Context.WasRollbackCalled)
	s.False(s.unitOfWork.Context.WasCommitCalled)
	s.False(s.unitOfWork.ReminderChannels().WasCreateCalled)
	s.False(s.unitOfWork.ReminderChannels().WasDeleteCalled)
}

func (s *testSuite) TestErrorReminderIsNotActive() {
	s.unitOfWork.Reminders().GetByIDReminder.Status = reminder.StatusSentSuccess

	_, err := s.service.Run(context.Background(), s.input)

	s.ErrorIs(err, reminder.ErrReminderNotActive)
	s.True(s.unitOfWork.Context.WasRollbackCalled)
	s.False(s.unitOfWork.Context.WasCommitCalled)
	s.False(s.unitOfWork.ReminderChannels().WasCreateCalled)
	s.False(s.unitOfWork.ReminderChannels().WasDeleteCalled)
}

func (s *testSuite) TestErrorChannelNotFound() {
	s.unitOfWork.Channels().ReadChannels = []channel.Channel{}

	_, err := s.service.Run(context.Background(), s.input)

	s.ErrorIs(err, reminder.ErrReminderChannelsNotValid)
	s.True(s.unitOfWork.Context.WasRollbackCalled)
	s.False(s.unitOfWork.Context.WasCommitCalled)
	s.False(s.unitOfWork.ReminderChannels().WasCreateCalled)
	s.False(s.unitOfWork.ReminderChannels().WasDeleteCalled)
}

func (s *testSuite) TestErrorChannelNotVerified() {
	s.unitOfWork.Channels().ReadChannels[1].VerifiedAt = c.NewOptional(Now, false)

	_, err := s.service.Run(context.Background(), s.input)

	s.ErrorIs(err, reminder.ErrReminderChannelsNotVerified)
	s.True(s.unitOfWork.Context.WasRollbackCalled)
	s.False(s.unitOfWork.Context.WasCommitCalled)
	s.False(s.unitOfWork.ReminderChannels().WasCreateCalled)
	s.False(s.unitOfWork.ReminderChannels().WasDeleteCalled)
}
