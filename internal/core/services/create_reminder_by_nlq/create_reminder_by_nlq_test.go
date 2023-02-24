package createreminderbynlq

import (
	"context"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	createreminder "remindme/internal/core/services/create_reminder"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	ReminderAt        = time.Date(2020, 6, 15, 15, 1, 1, 1, time.UTC)
	ChannelVerifiedAt = time.Date(2020, 1, 1, 1, 1, 1, 1, time.UTC)
)

type innerCreateService struct {
	result     createreminder.Result
	err        error
	calledWith []createreminder.Input
	lock       sync.Mutex
}

func NewInnerCreateService() *innerCreateService {
	return &innerCreateService{}
}

func (s *innerCreateService) Run(
	ctx context.Context,
	input createreminder.Input,
) (result createreminder.Result, err error) {
	if s.err != nil {
		return result, s.err
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.calledWith = append(s.calledWith, input)
	return s.result, nil
}

type suite struct {
	log          *logging.FakeLogger
	parser       *reminder.TestNLQParser
	channelRepo  *channel.FakeRepository
	now          time.Time
	innerService *innerCreateService
}

func setupSuite(now time.Time) *suite {
	parser := reminder.NewTestNLQParser()
	return &suite{
		log:          logging.NewFakeLogger(),
		parser:       parser,
		channelRepo:  channel.NewFakeRepository(),
		now:          now,
		innerService: NewInnerCreateService(),
	}
}

func (s *suite) createService() services.Service[Input, createreminder.Result] {
	return New(s.log, s.parser, s.channelRepo, func() time.Time { return s.now }, s.innerService)
}

func TestReminderCreatedSuccessfully(t *testing.T) {
	cases := []struct {
		id                 string
		userTimezone       *time.Location
		query              string
		now                time.Time
		readChannels       []channel.Channel
		expectedChannelIDs []channel.ID
	}{
		{
			id:           "1",
			userTimezone: time.UTC,
			query:        "15pm",
			now:          time.Date(2020, 1, 2, 3, 4, 5, 6, time.UTC),
			readChannels: []channel.Channel{
				{ID: 1, Type: channel.Email, VerifiedAt: c.NewOptional(ChannelVerifiedAt, true)},
			},
			expectedChannelIDs: []channel.ID{1},
		},
		{
			id:           "2",
			userTimezone: tz("Europe/Kaliningrad"),
			query:        "every 5 hours",
			now:          ReminderAt.Add(-time.Hour - time.Second),
			readChannels: []channel.Channel{
				{ID: 1, Type: channel.Email, VerifiedAt: c.NewOptional(ChannelVerifiedAt, true)},
				{ID: 2, Type: channel.Email, VerifiedAt: c.NewOptional(ChannelVerifiedAt, false)},
				{ID: 3, Type: channel.Telegram, VerifiedAt: c.NewOptional(ChannelVerifiedAt, true)},
				{ID: 4, Type: channel.Websocket, VerifiedAt: c.NewOptional(ChannelVerifiedAt, true)},
			},
			expectedChannelIDs: []channel.ID{1, 3},
		},
		{
			id:           "3",
			userTimezone: tz("Europe/Kaliningrad"),
			query:        "in 10h",
			now:          ReminderAt.Add(-time.Hour),
			readChannels: []channel.Channel{
				{ID: 1, Type: channel.Email, VerifiedAt: c.NewOptional(ChannelVerifiedAt, true)},
				{ID: 2, Type: channel.Email, VerifiedAt: c.NewOptional(ChannelVerifiedAt, false)},
				{ID: 3, Type: channel.Telegram, VerifiedAt: c.NewOptional(ChannelVerifiedAt, false)},
				{ID: 4, Type: channel.Websocket, VerifiedAt: c.NewOptional(ChannelVerifiedAt, true)},
			},
			expectedChannelIDs: []channel.ID{4},
		},
		{
			id:           "4",
			userTimezone: time.UTC,
			query:        "in 10h",
			now:          ReminderAt.Add(-72 * time.Hour),
			readChannels: []channel.Channel{
				{ID: 1, Type: channel.Email, VerifiedAt: c.NewOptional(ChannelVerifiedAt, false)},
				{ID: 2, Type: channel.Email, VerifiedAt: c.NewOptional(ChannelVerifiedAt, false)},
				{ID: 3, Type: channel.Telegram, VerifiedAt: c.NewOptional(ChannelVerifiedAt, false)},
				{ID: 4, Type: channel.Websocket, VerifiedAt: c.NewOptional(ChannelVerifiedAt, true)},
			},
			expectedChannelIDs: []channel.ID{4},
		},
		{
			id:           "5",
			userTimezone: time.UTC,
			query:        "in 10h",
			now:          ReminderAt.Add(-72 * time.Hour),
			readChannels: []channel.Channel{
				{ID: 100, Type: channel.Email, VerifiedAt: c.NewOptional(ChannelVerifiedAt, false)},
			},
			expectedChannelIDs: []channel.ID{},
		},
		{
			id:           "6",
			userTimezone: time.UTC,
			query:        "in 30m",
			now:          ReminderAt.Add(-30 * time.Minute),
			readChannels: []channel.Channel{
				{ID: 100, Type: channel.Email, VerifiedAt: c.NewOptional(ChannelVerifiedAt, true)},
			},
			expectedChannelIDs: []channel.ID{100},
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			// Setup ---
			suite := setupSuite(testcase.now)
			suite.parser.Params = reminder.CreateReminderParams{At: ReminderAt.In(testcase.userTimezone)}
			suite.channelRepo.ReadChannels = testcase.readChannels
			service := suite.createService()

			// Exercise ---
			input := Input{Query: testcase.query}
			input.User.ID = 1
			input.User.TimeZone = testcase.userTimezone
			_, err := service.Run(context.Background(), input)

			// Verify ---
			require.NoError(t, err)
			require.Len(t, suite.parser.CalledWith, 1)
			require.Equal(t, testcase.query, suite.parser.CalledWith[0].Query)
			require.Equal(t, testcase.now.In(testcase.userTimezone), suite.parser.CalledWith[0].UserLocalTime)
			require.Len(t, suite.innerService.calledWith, 1)
			require.Equal(t, ReminderAt.In(time.UTC), suite.innerService.calledWith[0].At)
			require.Equal(
				t,
				reminder.NewChannelIDs(testcase.expectedChannelIDs...),
				suite.innerService.calledWith[0].ChannelIDs,
			)
		})
	}
}

func TestReminderNLQParsingError(t *testing.T) {
	// Setup ---
	suite := setupSuite(ReminderAt.Add(-10 * time.Hour))
	suite.parser.ParseError = reminder.ErrNaturalQueryParsing
	service := suite.createService()

	// Exercise ---
	_, err := service.Run(
		context.Background(),
		Input{Query: "test", User: user.User{TimeZone: tz("Europe/Kaliningrad")}},
	)

	// Verify ---
	require.ErrorIs(t, err, reminder.ErrNaturalQueryParsing)
	require.Len(t, suite.innerService.calledWith, 0)
}

func tz(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}
