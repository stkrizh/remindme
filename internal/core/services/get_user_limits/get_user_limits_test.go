package getuserlimits

import (
	"context"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	NOW     = time.Date(2022, 6, 15, 12, 34, 55, 1, time.UTC)
	USER_ID = user.ID(1)
)

type Fixture struct {
	log       *logging.FakeLogger
	limits    *user.FakeLimitsRepository
	channels  *channel.FakeRepository
	reminders *reminder.TestReminderRepository
}

func NewFixture() Fixture {
	return Fixture{
		log:       logging.NewFakeLogger(),
		limits:    user.NewFakeLimitsRepository(),
		channels:  channel.NewFakeRepository(),
		reminders: reminder.NewTestReminderRepository(),
	}
}

func (f *Fixture) service() services.Service[Input, Result] {
	return New(f.log, f.limits, f.channels, f.reminders, func() time.Time { return NOW })
}

func TestGetUserLimitsSuccess(t *testing.T) {
	cases := []struct {
		id                   string
		limits               user.Limits
		channels             []channel.Channel
		reminderCount        uint32
		expectedActualLimits user.Limits
	}{
		{
			id:                   "1",
			limits:               user.Limits{},
			channels:             []channel.Channel{},
			reminderCount:        100,
			expectedActualLimits: user.Limits{},
		},
		{
			id:            "2",
			limits:        user.Limits{EmailChannelCount: c.NewOptional(uint32(2), true)},
			channels:      []channel.Channel{},
			reminderCount: 100,
			expectedActualLimits: user.Limits{
				EmailChannelCount: c.NewOptional(uint32(0), true),
			},
		},
		{
			id:            "3",
			limits:        user.Limits{TelegramChannelCount: c.NewOptional(uint32(2), true)},
			channels:      []channel.Channel{},
			reminderCount: 100,
			expectedActualLimits: user.Limits{
				TelegramChannelCount: c.NewOptional(uint32(0), true),
			},
		},
		{
			id: "4",
			limits: user.Limits{
				EmailChannelCount:    c.NewOptional(uint32(5), true),
				TelegramChannelCount: c.NewOptional(uint32(2), true),
			},
			channels: []channel.Channel{
				{Type: channel.Email},
				{Type: channel.Email},
				{Type: channel.Telegram},
				{Type: channel.Internal},
			},
			reminderCount: 100,
			expectedActualLimits: user.Limits{
				TelegramChannelCount: c.NewOptional(uint32(1), true),
				EmailChannelCount:    c.NewOptional(uint32(2), true),
			},
		},
		{
			id: "5",
			limits: user.Limits{
				EmailChannelCount: c.NewOptional(uint32(1), true),
			},
			channels: []channel.Channel{
				{Type: channel.Email},
				{Type: channel.Email},
				{Type: channel.Telegram},
				{Type: channel.Internal},
				{Type: channel.Email},
			},
			reminderCount: 100,
			expectedActualLimits: user.Limits{
				EmailChannelCount: c.NewOptional(uint32(3), true),
			},
		},
		{
			id: "6",
			limits: user.Limits{
				TelegramChannelCount: c.NewOptional(uint32(2), true),
			},
			channels: []channel.Channel{
				{Type: channel.Telegram},
				{Type: channel.Email},
				{Type: channel.Telegram},
				{Type: channel.Internal},
				{Type: channel.Telegram},
			},
			reminderCount: 100,
			expectedActualLimits: user.Limits{
				TelegramChannelCount: c.NewOptional(uint32(3), true),
			},
		},
		{
			id: "7",
			limits: user.Limits{
				ActiveReminderCount: c.NewOptional(uint32(5), true),
			},
			channels:      []channel.Channel{},
			reminderCount: 10,
			expectedActualLimits: user.Limits{
				ActiveReminderCount: c.NewOptional(uint32(10), true),
			},
		},
		{
			id: "8",
			limits: user.Limits{
				MonthlySentReminderCount: c.NewOptional(uint32(50), true),
			},
			channels:      []channel.Channel{},
			reminderCount: 40,
			expectedActualLimits: user.Limits{
				MonthlySentReminderCount: c.NewOptional(uint32(40), true),
			},
		},
		{
			id: "9",
			limits: user.Limits{
				EmailChannelCount:        c.NewOptional(uint32(1), true),
				TelegramChannelCount:     c.NewOptional(uint32(2), true),
				ActiveReminderCount:      c.NewOptional(uint32(10), true),
				MonthlySentReminderCount: c.NewOptional(uint32(50), true),
			},
			channels: []channel.Channel{
				{Type: channel.Internal},
				{Type: channel.Email},
				{Type: channel.Telegram},
			},
			reminderCount: 5,
			expectedActualLimits: user.Limits{
				EmailChannelCount:        c.NewOptional(uint32(1), true),
				TelegramChannelCount:     c.NewOptional(uint32(1), true),
				ActiveReminderCount:      c.NewOptional(uint32(5), true),
				MonthlySentReminderCount: c.NewOptional(uint32(5), true),
			},
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			fixture := NewFixture()
			fixture.limits.Limits = testcase.limits
			fixture.channels.ReadChannels = testcase.channels
			fixture.reminders.CountResult = uint(testcase.reminderCount)

			result, err := fixture.service().Run(context.Background(), Input{UserID: USER_ID})

			assert := require.New(t)
			assert.Nil(err)
			assert.Equal(testcase.limits, result.Limits)
			assert.Equal(testcase.expectedActualLimits, result.Values)
		})
	}
}

func TestGetUserLimitsChannelsRead(t *testing.T) {
	fixture := NewFixture()
	fixture.limits.Limits = user.Limits{EmailChannelCount: c.NewOptional(uint32(5), true)}

	_, err := fixture.service().Run(context.Background(), Input{UserID: USER_ID})

	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(
		[]channel.ReadOptions{{UserIDEquals: c.NewOptional(USER_ID, true)}},
		fixture.channels.Options,
	)
}

func TestGetUserLimitsRemindersCounted(t *testing.T) {
	fixture := NewFixture()
	fixture.limits.Limits = user.Limits{
		ActiveReminderCount:      c.NewOptional(uint32(5), true),
		MonthlySentReminderCount: c.NewOptional(uint32(10), true),
	}

	_, err := fixture.service().Run(context.Background(), Input{UserID: USER_ID})

	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(
		[]reminder.ReadOptions{
			{
				CreatedByEquals: c.NewOptional(USER_ID, true),
				StatusIn: c.NewOptional(
					[]reminder.Status{reminder.StatusCreated, reminder.StatusScheduled},
					true,
				),
			},
			{
				CreatedByEquals: c.NewOptional(USER_ID, true),
				StatusIn:        c.NewOptional([]reminder.Status{reminder.StatusSentSuccess}, true),
				SentAfter: c.NewOptional(
					time.Date(NOW.Year(), NOW.Month(), 1, 0, 0, 0, 0, time.UTC),
					true,
				),
			},
		},
		fixture.reminders.CountWith,
	)
}
