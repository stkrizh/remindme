package channel

import (
	"context"
	"reflect"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"remindme/internal/db"
	dbuser "remindme/internal/db/user"
	"testing"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/suite"
)

var NOW time.Time = time.Date(2020, 6, 6, 15, 30, 30, 0, time.UTC)

func TestChannelSettingsEncoding(t *testing.T) {
	cases := []struct {
		id       string
		chanType channel.Type
		settings channel.Settings
	}{
		{
			id:       "email-1",
			chanType: channel.Email,
			settings: &channel.EmailSettings{Email: c.NewEmail("test-1@test.com")},
		},
		{
			id:       "email-2",
			chanType: channel.Email,
			settings: &channel.EmailSettings{Email: c.NewEmail("test-2@test.com")},
		},
		{
			id:       "telegram-1",
			chanType: channel.Telegram,
			settings: &channel.TelegramSettings{
				Bot:    channel.TelegramBot("test"),
				ChatID: channel.TelegramChatID(1),
			},
		},
		{
			id:       "telegram-2",
			chanType: channel.Telegram,
			settings: &channel.TelegramSettings{
				Bot:    channel.TelegramBot("test-test-test-test aasdkhaskdhsakdhjksahdkjashd jkahsdkjas"),
				ChatID: channel.TelegramChatID(111222333444555),
			},
		},
		{
			id:       "ws-1",
			chanType: channel.Websocket,
			settings: &channel.WebsocketSettings{},
		},
	}
	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			encoded, err := encodeSettings(testcase.settings)
			if err != nil {
				t.Fatal("could not encode channel settings:", err)
			}

			decoded, err := decodeSettings(testcase.chanType, encoded)
			if err != nil {
				t.Fatal("could not decode channel settings:", err)
			}

			if !reflect.DeepEqual(testcase.settings, decoded) {
				t.Fatal("settings are not equal", testcase.settings, decoded)
			}
		})
	}
}

type testSuite struct {
	suite.Suite
	pool      *pgxpool.Pool
	repo      *PgxChannelRepository
	userRepo  *dbuser.PgxUserRepository
	user      user.User
	otherUser user.User
}

func (suite *testSuite) SetupSuite() {
	suite.pool = db.CreateTestPool()
	suite.repo = NewPgxChannelRepository(suite.pool)
	suite.userRepo = dbuser.NewPgxRepository(suite.pool)
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
			CreatedAt:    NOW,
			ActivatedAt:  c.NewOptional(NOW, true),
		},
	)
	s.Nil(err)
	s.user = u

	otherUser, err := s.userRepo.Create(
		context.Background(),
		user.CreateUserInput{
			Email:        c.NewOptional(c.NewEmail("test-2@test.test"), true),
			PasswordHash: c.NewOptional(user.PasswordHash("test"), true),
			CreatedAt:    NOW,
			ActivatedAt:  c.NewOptional(NOW, true),
		},
	)
	s.Nil(err)
	s.otherUser = otherUser
}

func (suite *testSuite) TearDownTest() {
	db.TruncateTables(suite.pool)
}

func TestPgxUserRepository(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestCreateSuccess() {
	type test struct {
		id    string
		input channel.CreateInput
	}
	cases := []test{
		{
			id: "email-1",
			input: channel.CreateInput{
				CreatedBy:         s.user.ID,
				Type:              channel.Email,
				Settings:          channel.NewEmailSettings("test-1@test.test"),
				CreatedAt:         time.Now().UTC().Truncate(time.Second),
				VerificationToken: c.NewOptional(channel.VerificationToken("test-1"), true),
			},
		},
		{
			id: "email-2",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Type:       channel.Email,
				Settings:   channel.NewEmailSettings("test-2@test.test"),
				CreatedAt:  NOW,
				VerifiedAt: c.NewOptional(NOW, true),
			},
		},
		{
			id: "telegram-1",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Type:       channel.Telegram,
				Settings:   channel.NewTelegramSettings("test-1", 1),
				CreatedAt:  time.Now().UTC().Truncate(time.Second),
				VerifiedAt: c.NewOptional(NOW, true),
			},
		},
		{
			id: "telegram-2",
			input: channel.CreateInput{
				CreatedBy:         s.user.ID,
				Type:              channel.Telegram,
				Settings:          channel.NewTelegramSettings("test-2", 111222333444),
				CreatedAt:         NOW,
				VerificationToken: c.NewOptional(channel.VerificationToken("test-2"), true),
			},
		},
		{
			id: "websocket-1",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Type:       channel.Websocket,
				Settings:   channel.NewWebsocketSettings(),
				CreatedAt:  time.Now().UTC().Truncate(time.Second),
				VerifiedAt: c.NewOptional(time.Now().UTC().Truncate(time.Second), true),
			},
		},
		{
			id: "websocket-2",
			input: channel.CreateInput{
				CreatedBy:         s.user.ID,
				Type:              channel.Websocket,
				Settings:          channel.NewWebsocketSettings(),
				CreatedAt:         NOW,
				VerificationToken: c.NewOptional(channel.VerificationToken("test-2"), true),
			},
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			newChannel, err := s.repo.Create(context.Background(), testcase.input)

			s.Nil(err)
			s.Equal(testcase.input.CreatedBy, newChannel.CreatedBy)
			s.Equal(testcase.input.Type, newChannel.Type)
			s.Equal(testcase.input.Settings, newChannel.Settings)
			s.Equal(testcase.input.CreatedAt, newChannel.CreatedAt)
			s.Equal(testcase.input.VerificationToken, newChannel.VerificationToken)
			s.Equal(testcase.input.VerifiedAt, newChannel.VerifiedAt)
		})
	}

}

func (s *testSuite) TestReadAndCount() {
	channelIDs := []channel.ID{
		s.createChannel(channel.Email, s.user),
		s.createChannel(channel.Email, s.otherUser),
		s.createChannel(channel.Telegram, s.user),
		s.createChannel(channel.Telegram, s.otherUser),
		s.createChannel(channel.Websocket, s.otherUser),
		s.createChannel(channel.Email, s.user),
		s.createChannel(channel.Telegram, s.otherUser),
	}

	cases := []struct {
		id          string
		options     channel.ReadOptions
		expectedIDs []channel.ID
	}{
		{
			id:          "1",
			options:     channel.ReadOptions{},
			expectedIDs: channelIDs,
		},
		{
			id:          "2",
			options:     channel.ReadOptions{UserIDEquals: c.NewOptional(s.user.ID, true)},
			expectedIDs: []channel.ID{channelIDs[0], channelIDs[2], channelIDs[5]},
		},
		{
			id:          "3",
			options:     channel.ReadOptions{UserIDEquals: c.NewOptional(s.otherUser.ID, true)},
			expectedIDs: []channel.ID{channelIDs[1], channelIDs[3], channelIDs[4], channelIDs[6]},
		},
		{
			id:          "4",
			options:     channel.ReadOptions{TypeEquals: c.NewOptional(channel.Email, true)},
			expectedIDs: []channel.ID{channelIDs[0], channelIDs[1], channelIDs[5]},
		},
		{
			id:          "5",
			options:     channel.ReadOptions{TypeEquals: c.NewOptional(channel.Telegram, true)},
			expectedIDs: []channel.ID{channelIDs[2], channelIDs[3], channelIDs[6]},
		},
		{
			id:          "6",
			options:     channel.ReadOptions{TypeEquals: c.NewOptional(channel.Websocket, true)},
			expectedIDs: []channel.ID{channelIDs[4]},
		},
		{
			id: "7",
			options: channel.ReadOptions{
				UserIDEquals: c.NewOptional(s.user.ID, true),
				TypeEquals:   c.NewOptional(channel.Email, true),
			},
			expectedIDs: []channel.ID{channelIDs[0], channelIDs[5]},
		},
		{
			id: "8",
			options: channel.ReadOptions{
				UserIDEquals: c.NewOptional(s.otherUser.ID, true),
				TypeEquals:   c.NewOptional(channel.Telegram, true),
			},
			expectedIDs: []channel.ID{channelIDs[3], channelIDs[6]},
		},
		{
			id: "9",
			options: channel.ReadOptions{
				UserIDEquals: c.NewOptional(s.user.ID, true),
				TypeEquals:   c.NewOptional(channel.Websocket, true),
			},
			expectedIDs: []channel.ID{},
		},
	}
	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			actualIDs := s.readChannelIDs(testcase.options)
			s.Truef(
				reflect.DeepEqual(testcase.expectedIDs, actualIDs),
				"expected: %v, actual: %v",
				testcase.expectedIDs,
				actualIDs,
			)
			actualCount := s.getCount(testcase.options)
			s.Equal(uint(len(testcase.expectedIDs)), actualCount)
		})
	}
}

func (s *testSuite) createChannel(t channel.Type, u user.User) channel.ID {
	s.T().Helper()

	var settings channel.Settings
	switch t {
	case channel.Email:
		settings = channel.NewEmailSettings(c.NewEmail("test@test.test"))
	case channel.Telegram:
		settings = channel.NewTelegramSettings(channel.TelegramBot("test"), channel.TelegramChatID(123))
	case channel.Websocket:
		settings = channel.NewWebsocketSettings()
	default:
		s.FailNow("unknown channel type", t)
	}
	createdChannel, err := s.repo.Create(
		context.Background(),
		channel.CreateInput{
			CreatedBy:         u.ID,
			Type:              t,
			Settings:          settings,
			CreatedAt:         NOW,
			VerificationToken: c.NewOptional(channel.VerificationToken("test-2"), true),
		},
	)
	s.Nil(err)
	return createdChannel.ID
}

func (s *testSuite) readChannelIDs(options channel.ReadOptions) []channel.ID {
	s.T().Helper()
	channels, err := s.repo.Read(context.Background(), options)
	s.Nil(err)
	channelIDs := make([]channel.ID, len(channels))
	for ix, c := range channels {
		channelIDs[ix] = c.ID
	}
	return channelIDs
}

func (s *testSuite) getCount(options channel.ReadOptions) uint {
	s.T().Helper()
	count, err := s.repo.Count(context.Background(), options)
	s.Nil(err)
	return uint(count)
}
