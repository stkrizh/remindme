package getlimitforchannels

import (
	"context"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
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
	log      *logging.FakeLogger
	limits   *user.FakeLimitsRepository
	channels *channel.FakeRepository
}

func NewFixture() Fixture {
	return Fixture{
		log:      logging.NewFakeLogger(),
		limits:   user.NewFakeLimitsRepository(),
		channels: channel.NewFakeRepository(),
	}
}

func (f *Fixture) service() services.Service[Input, Result] {
	return New(f.log, f.limits, f.channels)
}

func TestGetUserLimitsSuccess(t *testing.T) {
	cases := []struct {
		id             string
		limits         user.Limits
		channels       []channel.Channel
		expectedResult Result
	}{
		{
			id:             "1",
			limits:         user.Limits{},
			channels:       []channel.Channel{},
			expectedResult: Result{},
		},
		{
			id:       "2",
			limits:   user.Limits{EmailChannelCount: c.NewOptional(uint32(2), true)},
			channels: []channel.Channel{},
			expectedResult: Result{
				Email: c.NewOptional(user.Limit{Value: 2, Actual: 0}, true),
			},
		},
		{
			id:       "3",
			limits:   user.Limits{TelegramChannelCount: c.NewOptional(uint32(2), true)},
			channels: []channel.Channel{},
			expectedResult: Result{
				Telegram: c.NewOptional(user.Limit{Value: 2, Actual: 0}, true),
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
			expectedResult: Result{
				Email:    c.NewOptional(user.Limit{Value: 5, Actual: 2}, true),
				Telegram: c.NewOptional(user.Limit{Value: 2, Actual: 1}, true),
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
			expectedResult: Result{
				Email: c.NewOptional(user.Limit{Value: 1, Actual: 3}, true),
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
			expectedResult: Result{
				Telegram: c.NewOptional(user.Limit{Value: 2, Actual: 3}, true),
			},
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			fixture := NewFixture()
			fixture.limits.Limits = testcase.limits
			fixture.channels.ReadChannels = testcase.channels

			result, err := fixture.service().Run(context.Background(), Input{UserID: USER_ID})

			assert := require.New(t)
			assert.Nil(err)
			assert.Equal(testcase.expectedResult, result)
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
