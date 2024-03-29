package remidner

import (
	"context"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/db"
	dbchannel "remindme/internal/db/channel"
	dbuser "remindme/internal/db/user"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

const REMINDER_BODY = "test reminder body"

var (
	Now = time.Now().UTC()
	At  = Now.Add(time.Duration(time.Hour))
)

type testSuite struct {
	suite.Suite
	pool                *pgxpool.Pool
	repo                *PgxReminderRepository
	reminderChannelRepo *PgxReminderChannelRepository
	userRepo            *dbuser.PgxUserRepository
	channelRepo         *dbchannel.PgxChannelRepository
	user                user.User
	otherUser           user.User
	channel             channel.Channel
	otherChannel        channel.Channel
	otherUserChannel    channel.Channel
}

func (suite *testSuite) SetupSuite() {
	suite.pool = db.CreateTestPool()
	suite.repo = NewPgxReminderRepository(suite.pool)
	suite.reminderChannelRepo = NewPgxReminderChannelRepository(suite.pool)
	suite.userRepo = dbuser.NewPgxRepository(suite.pool)
	suite.channelRepo = dbchannel.NewPgxChannelRepository(suite.pool)
}

func (suite *testSuite) TearDownSuite() {
	suite.pool.Close()
}

func (s *testSuite) SetupTest() {
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

	otherUser, err := s.userRepo.Create(
		context.Background(),
		user.CreateUserInput{
			Email:        c.NewOptional(c.NewEmail("test-2@test.test"), true),
			PasswordHash: c.NewOptional(user.PasswordHash("test"), true),
			CreatedAt:    Now,
			ActivatedAt:  c.NewOptional(Now, true),
		},
	)
	s.Nil(err)
	s.otherUser = otherUser

	ch, err := s.channelRepo.Create(
		context.Background(),
		channel.CreateInput{
			CreatedBy:  u.ID,
			Type:       channel.Email,
			Settings:   channel.NewEmailSettings(c.NewEmail("test-1@test.test")),
			CreatedAt:  Now,
			VerifiedAt: c.NewOptional(Now, true),
		},
	)
	s.Nil(err)
	s.channel = ch
	ch, err = s.channelRepo.Create(
		context.Background(),
		channel.CreateInput{
			CreatedBy:  u.ID,
			Type:       channel.Email,
			Settings:   channel.NewEmailSettings(c.NewEmail("test-2@test.test")),
			CreatedAt:  Now,
			VerifiedAt: c.NewOptional(Now, true),
		},
	)
	s.Nil(err)
	s.otherChannel = ch

	ch, err = s.channelRepo.Create(
		context.Background(),
		channel.CreateInput{
			CreatedBy:  otherUser.ID,
			Type:       channel.Email,
			Settings:   channel.NewEmailSettings(c.NewEmail("test@test.test")),
			CreatedAt:  Now,
			VerifiedAt: c.NewOptional(Now, true),
		},
	)
	s.Nil(err)
	s.otherUserChannel = ch
}

func (suite *testSuite) TearDownTest() {
	db.TruncateTables(suite.pool)
}

func TestPgxReminderRepositories(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestCreateReminder() {
	cases := []struct {
		id    string
		input reminder.CreateInput
	}{
		{
			id: "1",
			input: reminder.CreateInput{
				CreatedBy: s.user.ID,
				CreatedAt: Now.Truncate(time.Millisecond),
				At:        Now.Truncate(time.Millisecond),
				Status:    reminder.StatusCreated,
			},
		},
		{
			id: "2",
			input: reminder.CreateInput{
				CreatedBy:   s.otherUser.ID,
				CreatedAt:   Now.Truncate(time.Millisecond),
				At:          time.Date(2023, 1, 15, 15, 31, 32, 0, time.UTC),
				Status:      reminder.StatusScheduled,
				ScheduledAt: c.NewOptional(time.Date(2023, 1, 15, 16, 31, 32, 0, time.UTC), true),
				Every:       c.NewOptional(reminder.EveryDay, true),
				Body:        "test-1",
			},
		},
		{
			id: "3",
			input: reminder.CreateInput{
				CreatedBy:   s.otherUser.ID,
				CreatedAt:   time.Date(2023, 12, 1, 10, 10, 10, 0, time.UTC),
				At:          time.Date(2023, 12, 1, 23, 23, 23, 0, time.UTC),
				Status:      reminder.StatusScheduled,
				ScheduledAt: c.NewOptional(time.Date(2023, 12, 1, 10, 10, 10, 0, time.UTC), true),
				Every:       c.NewOptional(reminder.NewEvery(3, reminder.PeriodMonth), true),
				Body:        "test-2",
			},
		},
	}

	for _, testcase := range cases {
		reminder, err := s.repo.Create(context.Background(), testcase.input)

		assert := s.Require()
		assert.Nil(err, testcase.id)
		assert.True(reminder.ID > 0, testcase.id)
		assert.Equal(testcase.input.CreatedBy, reminder.CreatedBy, testcase.id)
		assert.Equal(testcase.input.At, reminder.At, testcase.id)
		assert.Equal(testcase.input.Every, reminder.Every, testcase.id)
		assert.Equal(testcase.input.Status, reminder.Status, testcase.id)
		assert.Equal(testcase.input.ScheduledAt, reminder.ScheduledAt, testcase.id)
		assert.Equal(testcase.input.Body, reminder.Body, testcase.id)
		assert.False(reminder.SentAt.IsPresent, testcase.id)
		assert.False(reminder.CanceledAt.IsPresent, testcase.id)
	}
}

func (s *testSuite) TestCreateReminderChannels() {
	cases := []struct {
		id    string
		input reminder.CreateChannelsInput
	}{
		{
			id: "1",
			input: reminder.CreateChannelsInput{
				ChannelIDs: reminder.NewChannelIDs(),
			},
		},
		{
			id: "2",
			input: reminder.CreateChannelsInput{
				ChannelIDs: reminder.NewChannelIDs(s.channel.ID),
			},
		},
		{
			id: "3",
			input: reminder.CreateChannelsInput{
				ChannelIDs: reminder.NewChannelIDs(s.channel.ID, s.otherChannel.ID),
			},
		},
	}

	for _, testcase := range cases {
		reminder := s.createReminder()
		testcase.input.ReminderID = reminder.ID
		channelIDs, err := s.reminderChannelRepo.Create(context.Background(), testcase.input)

		assert := s.Require()
		assert.Nil(err, testcase.id)
		assert.Equal(testcase.input.ChannelIDs, channelIDs, testcase.id)
	}
}

func (s *testSuite) TestGetByID() {
	expectedReminder1 := s.createReminder()
	_, err := s.reminderChannelRepo.Create(
		context.Background(),
		reminder.NewCreateChannelsInput(expectedReminder1.ID, s.channel.ID),
	)
	s.Nil(err)

	expectedReminder2 := s.createReminder()
	_, err = s.reminderChannelRepo.Create(
		context.Background(),
		reminder.NewCreateChannelsInput(expectedReminder2.ID, s.channel.ID, s.otherChannel.ID),
	)
	s.Nil(err)

	reminder1, err := s.repo.GetByID(context.Background(), expectedReminder1.ID)
	s.Nil(err)
	s.Equal(expectedReminder1, reminder1.Reminder)
	s.Equal([]channel.ID{s.channel.ID}, reminder1.ChannelIDs)

	reminder2, err := s.repo.GetByID(context.Background(), expectedReminder2.ID)
	s.Nil(err)
	s.Equal(expectedReminder2, reminder2.Reminder)
	s.Equal([]channel.ID{s.channel.ID, s.otherChannel.ID}, reminder2.ChannelIDs)

	_, err = s.repo.GetByID(context.Background(), reminder.ID(111222333))
	s.ErrorIs(err, reminder.ErrReminderDoesNotExist)
}

func (s *testSuite) TestReadAndCount() {
	reminderIDs := s.createReminders([]reminder.CreateInput{
		{
			// 0
			CreatedBy: s.user.ID,
			CreatedAt: Now,
			At:        At.Add(time.Hour),
			Status:    reminder.StatusCreated,
		},
		{
			// 1
			CreatedBy:   s.user.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			Status:      reminder.StatusScheduled,
		},
		{
			// 2
			CreatedBy:   s.user.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			SentAt:      c.NewOptional(At.Add(time.Second), true),
			Status:      reminder.StatusSentSuccess,
		},
		{
			// 3
			CreatedBy:   s.user.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			SentAt:      c.NewOptional(At.Add(time.Second), true),
			Status:      reminder.StatusSentError,
		},
		{
			// 4
			CreatedBy:  s.user.ID,
			CreatedAt:  Now,
			At:         At,
			CanceledAt: c.NewOptional(Now.Add(time.Second), true),
			Status:     reminder.StatusCanceled,
		},
		{
			// 5
			CreatedBy: s.user.ID,
			CreatedAt: Now,
			At:        At,
			SentAt:    c.NewOptional(At.Add(time.Second), true),
			Status:    reminder.StatusSentLimitExceeded,
		},
		{
			// 6
			CreatedBy: s.otherUser.ID,
			CreatedAt: Now,
			At:        At.Add(-time.Hour),
			Status:    reminder.StatusCreated,
		},
		{
			// 7
			CreatedBy:   s.otherUser.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			Status:      reminder.StatusScheduled,
		},
		{
			// 8
			CreatedBy:   s.otherUser.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			SentAt:      c.NewOptional(At.Add(time.Second), true),
			Status:      reminder.StatusSentSuccess,
		},
		{
			// 9
			CreatedBy:   s.otherUser.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			SentAt:      c.NewOptional(At.Add(2*time.Second), true),
			Status:      reminder.StatusSentSuccess,
		},
		{
			// 10
			CreatedBy:   s.otherUser.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			SentAt:      c.NewOptional(At.Add(time.Second), true),
			Status:      reminder.StatusSentError,
		},
		{
			// 11
			CreatedBy:  s.otherUser.ID,
			CreatedAt:  Now,
			At:         At,
			CanceledAt: c.NewOptional(Now.Add(time.Second), true),
			Status:     reminder.StatusCanceled,
		},
		{
			// 12
			CreatedBy:   s.otherUser.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			SentAt:      c.NewOptional(At.Add(time.Second), true),
			Status:      reminder.StatusSentLimitExceeded,
		},
		{
			// 13
			CreatedBy:   s.otherUser.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			Status:      reminder.StatusSending,
		},
	})

	cases := []struct {
		id            string
		options       reminder.ReadOptions
		expectedIxs   []int
		expectedCount uint
	}{
		{
			id:            "1",
			options:       reminder.ReadOptions{},
			expectedIxs:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13},
			expectedCount: 14,
		},
		{
			id:            "2",
			options:       reminder.ReadOptions{OrderBy: reminder.OrderByIDDesc},
			expectedIxs:   []int{13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
			expectedCount: 14,
		},
		{
			id:            "3",
			options:       reminder.ReadOptions{OrderBy: reminder.OrderByAtAsc},
			expectedIxs:   []int{6, 1, 2, 3, 4, 5, 7, 8, 9, 10, 11, 12, 13, 0},
			expectedCount: 14,
		},
		{
			id:            "4",
			options:       reminder.ReadOptions{OrderBy: reminder.OrderByAtDesc},
			expectedIxs:   []int{0, 1, 2, 3, 4, 5, 7, 8, 9, 10, 11, 12, 13, 6},
			expectedCount: 14,
		},
		{
			id: "5",
			options: reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(s.user.ID, true),
			},
			expectedIxs:   []int{0, 1, 2, 3, 4, 5},
			expectedCount: 6,
		},
		{
			id: "6",
			options: reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(s.otherUser.ID, true),
				OrderBy:         reminder.OrderByIDDesc,
			},
			expectedIxs:   []int{13, 12, 11, 10, 9, 8, 7, 6},
			expectedCount: 8,
		},
		{
			id: "7",
			options: reminder.ReadOptions{
				SentAfter: c.NewOptional(At.Add(3*time.Second), true),
				OrderBy:   reminder.OrderByIDDesc,
			},
			expectedIxs:   []int{},
			expectedCount: 0,
		},
		{
			id: "8",
			options: reminder.ReadOptions{
				SentAfter: c.NewOptional(At, true),
			},
			expectedIxs:   []int{2, 3, 5, 8, 9, 10, 12},
			expectedCount: 7,
		},
		{
			id: "9",
			options: reminder.ReadOptions{
				SentAfter: c.NewOptional(At, true),
				StatusIn:  c.NewOptional([]reminder.Status{reminder.StatusSentSuccess}, true),
			},
			expectedIxs:   []int{2, 8, 9},
			expectedCount: 3,
		},
		{
			id: "10",
			options: reminder.ReadOptions{
				SentAfter:       c.NewOptional(At, true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSentSuccess}, true),
				CreatedByEquals: c.NewOptional(s.user.ID, true),
			},
			expectedIxs:   []int{2},
			expectedCount: 1,
		},
		{
			id: "11",
			options: reminder.ReadOptions{
				SentAfter:       c.NewOptional(At.Add(2*time.Second), true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSentSuccess}, true),
				CreatedByEquals: c.NewOptional(s.otherUser.ID, true),
			},
			expectedIxs:   []int{9},
			expectedCount: 1,
		},
		{
			id: "12",
			options: reminder.ReadOptions{
				SentAfter:       c.NewOptional(At, true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSentSuccess}, true),
				CreatedByEquals: c.NewOptional(s.otherUser.ID, true),
				OrderBy:         reminder.OrderByIDDesc,
			},
			expectedIxs:   []int{9, 8},
			expectedCount: 2,
		},
		{
			id: "13",
			options: reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(s.user.ID, true),
				Limit:           c.NewOptional(uint(2), true),
				Offset:          0,
			},
			expectedIxs:   []int{0, 1},
			expectedCount: 6,
		},
		{
			id: "14",
			options: reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(s.user.ID, true),
				Limit:           c.NewOptional(uint(3), true),
				Offset:          2,
			},
			expectedIxs:   []int{2, 3, 4},
			expectedCount: 6,
		},
		{
			id: "15",
			options: reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(s.otherUser.ID, true),
				OrderBy:         reminder.OrderByIDDesc,
				StatusIn: c.NewOptional([]reminder.Status{
					reminder.StatusCanceled,
					reminder.StatusSentSuccess,
					reminder.StatusScheduled,
				}, true),
				Limit:  c.NewOptional(uint(2), true),
				Offset: 1,
			},
			expectedIxs:   []int{9, 8},
			expectedCount: 4,
		},
		{
			id: "16",
			options: reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(s.otherUser.ID, true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSending}, true),
			},
			expectedIxs:   []int{13},
			expectedCount: 1,
		},
		{
			id: "17",
			options: reminder.ReadOptions{
				CreatedByEquals: c.NewOptional(s.user.ID, true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSending}, true),
			},
			expectedIxs:   []int{},
			expectedCount: 0,
		},
	}
	for _, testcase := range cases {
		reminders, err := s.repo.Read(context.Background(), testcase.options)
		s.Nil(err, testcase.id)
		s.assertReminderIDsEqual(testcase.id, reminderIDs, testcase.expectedIxs, reminders)

		count, err := s.repo.Count(context.Background(), testcase.options)
		s.Nil(err, testcase.id)
		s.Equal(testcase.expectedCount, count, testcase.id)
	}
}

func (s *testSuite) TestReadReminderChannels() {
	r1 := s.createReminder()
	_, err := s.reminderChannelRepo.Create(context.Background(), reminder.NewCreateChannelsInput(r1.ID, s.channel.ID))
	s.Nil(err)

	r2 := s.createReminder()
	_, err = s.reminderChannelRepo.Create(
		context.Background(),
		reminder.NewCreateChannelsInput(r2.ID, s.channel.ID, s.otherChannel.ID),
	)
	s.Nil(err)

	r3 := s.createReminder()
	_, err = s.reminderChannelRepo.Create(
		context.Background(),
		reminder.NewCreateChannelsInput(r3.ID, s.otherUserChannel.ID, s.channel.ID, s.otherChannel.ID),
	)
	s.Nil(err)

	reminders, err := s.repo.Read(context.Background(), reminder.ReadOptions{OrderBy: reminder.OrderByIDAsc})
	s.Nil(err)

	assert := s.Require()
	assert.Equal(3, len(reminders))
	assert.Equal([]channel.ID{s.channel.ID}, reminders[0].ChannelIDs)
	assert.Equal([]channel.ID{s.channel.ID, s.otherChannel.ID}, reminders[1].ChannelIDs)
	assert.Equal([]channel.ID{s.channel.ID, s.otherChannel.ID, s.otherUserChannel.ID}, reminders[2].ChannelIDs)
}

func (s *testSuite) TestDeleteReminderChannels() {
	rem := s.createReminderWithChannels()

	err := s.reminderChannelRepo.DeleteByReminderID(context.Background(), rem.ID)
	s.Nil(err)

	_, err = s.repo.GetByID(context.Background(), rem.ID)
	// ErrReminderDoesNotExist is returned due to the reminder does not have channels
	s.ErrorIs(err, reminder.ErrReminderDoesNotExist)
}

func (s *testSuite) TestUpdateSuccess() {
	cases := []struct {
		id    string
		input reminder.UpdateInput
	}{
		{
			id:    "1",
			input: reminder.UpdateInput{},
		},
		{
			id: "2",
			input: reminder.UpdateInput{
				DoAtUpdate: true,
				At:         time.Date(2000, 5, 12, 23, 13, 44, 0, time.UTC),
			},
		},
		{
			id: "3",
			input: reminder.UpdateInput{
				DoStatusUpdate:     true,
				Status:             reminder.StatusCanceled,
				DoCanceledAtUpdate: true,
				CanceledAt:         c.NewOptional(time.Date(2000, 5, 6, 7, 8, 9, 0, time.UTC), true),
			},
		},
		{
			id: "4",
			input: reminder.UpdateInput{
				DoStatusUpdate:      true,
				Status:              reminder.StatusScheduled,
				DoScheduledAtUpdate: true,
				ScheduledAt:         c.NewOptional(time.Date(2000, 6, 7, 8, 9, 10, 0, time.UTC), true),
			},
		},
		{
			id: "5",
			input: reminder.UpdateInput{
				DoStatusUpdate: true,
				Status:         reminder.StatusSentSuccess,
				DoSentAtUpdate: true,
				SentAt:         c.NewOptional(time.Date(2000, 7, 8, 9, 10, 11, 0, time.UTC), true),
			},
		},
		{
			id: "6",
			input: reminder.UpdateInput{
				DoStatusUpdate: true,
				Status:         reminder.StatusSentError,
				DoSentAtUpdate: true,
				SentAt:         c.NewOptional(time.Date(2000, 8, 9, 10, 11, 12, 0, time.UTC), true),
			},
		},
		{
			id: "7",
			input: reminder.UpdateInput{
				DoStatusUpdate: true,
				Status:         reminder.StatusSentLimitExceeded,
				DoSentAtUpdate: true,
				SentAt:         c.NewOptional(time.Date(2000, 10, 11, 12, 13, 14, 0, time.UTC), true),
			},
		},
		{
			id: "8",
			input: reminder.UpdateInput{
				DoEveryUpdate: true,
				Every:         c.NewOptional(reminder.NewEvery(3, reminder.PeriodDay), true),
			},
		},
		{
			id: "9",
			input: reminder.UpdateInput{
				DoEveryUpdate: true,
				Every:         c.NewOptional(reminder.EveryYear, true),
			},
		},
		{
			id: "10",
			input: reminder.UpdateInput{
				DoEveryUpdate: true,
			},
		},
		{
			id: "11",
			input: reminder.UpdateInput{
				DoBodyUpdate: true,
			},
		},
		{
			id: "12",
			input: reminder.UpdateInput{
				DoBodyUpdate: true,
				Body:         "test new reminder body",
			},
		},
		{
			id: "13",
			input: reminder.UpdateInput{
				DoStatusUpdate: true,
				Status:         reminder.StatusSending,
			},
		},
	}

	for _, testcase := range cases {
		reminderBefore := s.createReminderWithChannels()
		testcase.input.ID = reminderBefore.ID
		rem, err := s.repo.Update(context.Background(), testcase.input)

		assert := s.Require()
		assert.Nil(err, testcase.id)
		if testcase.input.DoAtUpdate {
			assert.Equal(testcase.input.At, rem.At, testcase.id)
		} else {
			assert.Equal(reminderBefore.At, rem.At, testcase.id)
		}
		if testcase.input.DoBodyUpdate {
			assert.Equal(testcase.input.Body, rem.Body, testcase.id)
		} else {
			assert.Equal(reminderBefore.Body, rem.Body, testcase.id)
		}
		if testcase.input.DoEveryUpdate {
			assert.Equal(testcase.input.Every, rem.Every, testcase.id)
		} else {
			assert.Equal(reminderBefore.Every, rem.Every, testcase.id)
		}
		if testcase.input.DoStatusUpdate {
			assert.Equal(testcase.input.Status, rem.Status, testcase.id)
		} else {
			assert.Equal(reminderBefore.Status, rem.Status, testcase.id)
		}
		if testcase.input.DoScheduledAtUpdate {
			assert.Equal(testcase.input.ScheduledAt, rem.ScheduledAt, testcase.id)
		} else {
			assert.Equal(reminderBefore.ScheduledAt, rem.ScheduledAt, testcase.id)
		}
		if testcase.input.DoSentAtUpdate {
			assert.Equal(testcase.input.SentAt, rem.SentAt, testcase.id)
		} else {
			assert.Equal(reminderBefore.SentAt, rem.SentAt, testcase.id)
		}
		if testcase.input.DoCanceledAtUpdate {
			assert.Equal(testcase.input.CanceledAt, rem.CanceledAt, testcase.id)
		} else {
			assert.Equal(reminderBefore.CanceledAt, rem.CanceledAt, testcase.id)
		}

		remAfter, err := s.repo.GetByID(context.Background(), rem.ID)
		assert.Nil(err, testcase.id)
		assert.Equal(rem, remAfter.Reminder, testcase.id)
	}
}

func (s *testSuite) TestDeleteSuccess() {
	// Setup ---
	rem := s.createReminder()

	// Exercise ---
	err := s.repo.Delete(context.Background(), rem.ID)

	// Verify ---
	s.Nil(err)
	_, err = s.repo.GetByID(context.Background(), rem.ID)
	s.ErrorIs(err, reminder.ErrReminderDoesNotExist)
}

func (s *testSuite) TestDeleteNotExisting() {
	// Exercise ---
	err := s.repo.Delete(context.Background(), reminder.ID(111222333))

	// Verify ---
	s.ErrorIs(err, reminder.ErrReminderDoesNotExist)
}

func (s *testSuite) createReminder() reminder.Reminder {
	s.T().Helper()
	r, err := s.repo.Create(
		context.Background(),
		reminder.CreateInput{
			CreatedBy: s.user.ID,
			At:        Now.Add(time.Duration(time.Hour * 3)),
			CreatedAt: Now,
			Status:    reminder.StatusCreated,
			Body:      REMINDER_BODY,
		},
	)
	s.Nil(err)
	return r
}

func (s *testSuite) createReminderWithChannels() (rem reminder.ReminderWithChannels) {
	s.T().Helper()
	r := s.createReminder()
	_, err := s.reminderChannelRepo.Create(
		context.Background(),
		reminder.NewCreateChannelsInput(r.ID, s.channel.ID, s.otherChannel.ID),
	)
	s.Nil(err)
	rem.FromReminderAndChannels(r, []channel.ID{s.channel.ID, s.otherChannel.ID})
	return rem
}

func (s *testSuite) createReminders(inputs []reminder.CreateInput) []reminder.ID {
	s.T().Helper()
	reminderIDs := make([]reminder.ID, 0, len(inputs))
	for _, input := range inputs {
		rem, err := s.repo.Create(context.Background(), input)
		if err != nil {
			s.FailNow("could not create reminder: %v, %v", input, err)
		}
		reminderIDs = append(reminderIDs, rem.ID)
		_, err = s.reminderChannelRepo.Create(
			context.Background(),
			reminder.NewCreateChannelsInput(rem.ID, s.channel.ID, s.otherChannel.ID),
		)
		if err != nil {
			s.FailNow("could not create reminder channels: %v, %v", input, err)
		}
	}
	return reminderIDs
}

func (s *testSuite) assertReminderIDsEqual(
	testacaseID string,
	reminderIDs []reminder.ID,
	expectedIxs []int,
	actualReminders []reminder.ReminderWithChannels,
) {
	s.T().Helper()

	expectedIDs := make([]reminder.ID, 0, len(expectedIxs))
	for _, expectedIx := range expectedIxs {
		expectedIDs = append(expectedIDs, reminderIDs[expectedIx])
	}

	actualIDs := make([]reminder.ID, 0, len(actualReminders))
	for _, rem := range actualReminders {
		actualIDs = append(actualIDs, rem.ID)
	}

	s.Equal(expectedIDs, actualIDs, testacaseID)
}

func dt(date string) time.Time {
	t, err := time.Parse(time.RFC3339, date)
	if err != nil {
		panic(date)
	}
	return t
}
