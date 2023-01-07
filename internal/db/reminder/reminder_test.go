package remidner

import (
	"context"
	"reflect"
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
			},
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			reminder, err := s.repo.Create(context.Background(), testcase.input)

			assert := s.Require()
			assert.Nil(err)
			assert.True(reminder.ID > 0)
			assert.Equal(testcase.input.CreatedBy, reminder.CreatedBy)
			assert.Equal(testcase.input.At, reminder.At)
			assert.Equal(testcase.input.Every, reminder.Every)
			assert.Equal(testcase.input.Status, reminder.Status)
			assert.Equal(testcase.input.ScheduledAt, reminder.ScheduledAt)
			assert.False(reminder.SentAt.IsPresent)
			assert.False(reminder.CanceledAt.IsPresent)
		})
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
		s.Run(testcase.id, func() {
			reminder := s.createReminder()
			testcase.input.ReminderID = reminder.ID
			channelIDs, err := s.reminderChannelRepo.Create(context.Background(), testcase.input)

			assert := s.Require()
			assert.Nil(err)
			assert.Equal(testcase.input.ChannelIDs, channelIDs)
		})
	}
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
			Status:      reminder.StatusSendSuccess,
		},
		{
			// 3
			CreatedBy:   s.user.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			SentAt:      c.NewOptional(At.Add(time.Second), true),
			Status:      reminder.StatusSendError,
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
			Status:    reminder.StatusSendLimitExceeded,
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
			Status:      reminder.StatusSendSuccess,
		},
		{
			// 9
			CreatedBy:   s.otherUser.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			SentAt:      c.NewOptional(At.Add(2*time.Second), true),
			Status:      reminder.StatusSendSuccess,
		},
		{
			// 10
			CreatedBy:   s.otherUser.ID,
			CreatedAt:   Now,
			At:          At,
			ScheduledAt: c.NewOptional(Now, true),
			SentAt:      c.NewOptional(At.Add(time.Second), true),
			Status:      reminder.StatusSendError,
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
			Status:      reminder.StatusSendLimitExceeded,
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
			expectedIxs:   []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			expectedCount: 13,
		},
		{
			id:            "2",
			options:       reminder.ReadOptions{OrderBy: reminder.OrderByIDDesc},
			expectedIxs:   []int{12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0},
			expectedCount: 13,
		},
		{
			id:            "3",
			options:       reminder.ReadOptions{OrderBy: reminder.OrderByAtAsc},
			expectedIxs:   []int{6, 1, 2, 3, 4, 5, 7, 8, 9, 10, 11, 12, 0},
			expectedCount: 13,
		},
		{
			id:            "4",
			options:       reminder.ReadOptions{OrderBy: reminder.OrderByAtDesc},
			expectedIxs:   []int{0, 1, 2, 3, 4, 5, 7, 8, 9, 10, 11, 12, 6},
			expectedCount: 13,
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
			expectedIxs:   []int{12, 11, 10, 9, 8, 7, 6},
			expectedCount: 7,
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
				StatusIn:  c.NewOptional([]reminder.Status{reminder.StatusSendSuccess}, true),
			},
			expectedIxs:   []int{2, 8, 9},
			expectedCount: 3,
		},
		{
			id: "10",
			options: reminder.ReadOptions{
				SentAfter:       c.NewOptional(At, true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSendSuccess}, true),
				CreatedByEquals: c.NewOptional(s.user.ID, true),
			},
			expectedIxs:   []int{2},
			expectedCount: 1,
		},
		{
			id: "11",
			options: reminder.ReadOptions{
				SentAfter:       c.NewOptional(At.Add(2*time.Second), true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSendSuccess}, true),
				CreatedByEquals: c.NewOptional(s.otherUser.ID, true),
			},
			expectedIxs:   []int{9},
			expectedCount: 1,
		},
		{
			id: "12",
			options: reminder.ReadOptions{
				SentAfter:       c.NewOptional(At, true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSendSuccess}, true),
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
					reminder.StatusSendSuccess,
					reminder.StatusScheduled,
				}, true),
				Limit:  c.NewOptional(uint(2), true),
				Offset: 1,
			},
			expectedIxs:   []int{9, 8},
			expectedCount: 4,
		},
	}
	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			reminders, err := s.repo.Read(context.Background(), testcase.options)
			s.Nil(err)
			s.assertReminderIDsEqual(reminderIDs, testcase.expectedIxs, reminders)

			count, err := s.repo.Count(context.Background(), testcase.options)
			s.Nil(err)
			s.Equal(testcase.expectedCount, count)
		})
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
	assert.True(reflect.DeepEqual([]channel.ID{s.channel.ID}, reminders[0].ChannelIDs))
	assert.True(reflect.DeepEqual([]channel.ID{s.channel.ID, s.otherChannel.ID}, reminders[1].ChannelIDs))
	assert.True(reflect.DeepEqual(
		[]channel.ID{s.channel.ID, s.otherChannel.ID, s.otherUserChannel.ID},
		reminders[2].ChannelIDs,
	))
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
		},
	)
	s.Nil(err)
	return r
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

	s.True(reflect.DeepEqual(expectedIDs, actualIDs), "not equal: %v, %v", expectedIDs, actualIDs)
}
