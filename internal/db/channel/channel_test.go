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

var (
	Now       time.Time = time.Date(2020, 6, 6, 15, 30, 30, 0, time.UTC)
	OtherTime time.Time = time.Date(2021, 1, 3, 4, 31, 32, 0, time.UTC)
)

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
}

func (suite *testSuite) TearDownTest() {
	db.TruncateTables(suite.pool)
}

func TestPgxChannelRepository(t *testing.T) {
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
				IsDefault:         true,
			},
		},
		{
			id: "email-2",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Type:       channel.Email,
				Settings:   channel.NewEmailSettings("test-2@test.test"),
				CreatedAt:  Now,
				VerifiedAt: c.NewOptional(Now, true),
			},
		},
		{
			id: "telegram-1",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Type:       channel.Telegram,
				Settings:   channel.NewTelegramSettings("test-1", 1),
				CreatedAt:  time.Now().UTC().Truncate(time.Second),
				VerifiedAt: c.NewOptional(Now, true),
				IsDefault:  true,
			},
		},
		{
			id: "telegram-2",
			input: channel.CreateInput{
				CreatedBy:         s.user.ID,
				Type:              channel.Telegram,
				Settings:          channel.NewTelegramSettings("test-2", 111222333444),
				CreatedAt:         Now,
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
				IsDefault:  true,
			},
		},
		{
			id: "websocket-2",
			input: channel.CreateInput{
				CreatedBy:         s.user.ID,
				Type:              channel.Websocket,
				Settings:          channel.NewWebsocketSettings(),
				CreatedAt:         Now,
				VerificationToken: c.NewOptional(channel.VerificationToken("test-2"), true),
			},
		},
	}

	for _, testcase := range cases {
		newChannel, err := s.repo.Create(context.Background(), testcase.input)

		s.Nil(err, testcase.id)
		s.Equal(testcase.input.CreatedBy, newChannel.CreatedBy, testcase.id)
		s.Equal(testcase.input.Type, newChannel.Type, testcase.id)
		s.Equal(testcase.input.Settings, newChannel.Settings, testcase.id)
		s.Equal(testcase.input.CreatedAt, newChannel.CreatedAt, testcase.id)
		s.Equal(testcase.input.VerificationToken, newChannel.VerificationToken, testcase.id)
		s.Equal(testcase.input.VerifiedAt, newChannel.VerifiedAt, testcase.id)
		s.Equal(testcase.input.IsDefault, newChannel.IsDefault, testcase.id)
	}

}

func (s *testSuite) TestReadAndCount() {
	channelIDs := []channel.ID{
		s.createChannel(channel.Email, s.user, true),
		s.createChannel(channel.Email, s.otherUser, false),
		s.createChannel(channel.Telegram, s.user, false),
		s.createChannel(channel.Telegram, s.otherUser, true),
		s.createChannel(channel.Websocket, s.otherUser, false),
		s.createChannel(channel.Email, s.user, false),
		s.createChannel(channel.Telegram, s.otherUser, false),
	}

	cases := []struct {
		id            string
		options       channel.ReadOptions
		expectedIDs   []channel.ID
		expectedCount uint
	}{
		{
			id:            "1",
			options:       channel.ReadOptions{},
			expectedIDs:   channelIDs,
			expectedCount: 7,
		},
		{
			id:            "2",
			options:       channel.ReadOptions{UserIDEquals: c.NewOptional(s.user.ID, true)},
			expectedIDs:   []channel.ID{channelIDs[0], channelIDs[2], channelIDs[5]},
			expectedCount: 3,
		},
		{
			id:            "3",
			options:       channel.ReadOptions{UserIDEquals: c.NewOptional(s.otherUser.ID, true)},
			expectedIDs:   []channel.ID{channelIDs[1], channelIDs[3], channelIDs[4], channelIDs[6]},
			expectedCount: 4,
		},
		{
			id:            "4",
			options:       channel.ReadOptions{TypeEquals: c.NewOptional(channel.Email, true)},
			expectedIDs:   []channel.ID{channelIDs[0], channelIDs[1], channelIDs[5]},
			expectedCount: 3,
		},
		{
			id:            "5",
			options:       channel.ReadOptions{TypeEquals: c.NewOptional(channel.Telegram, true)},
			expectedIDs:   []channel.ID{channelIDs[2], channelIDs[3], channelIDs[6]},
			expectedCount: 3,
		},
		{
			id:            "6",
			options:       channel.ReadOptions{TypeEquals: c.NewOptional(channel.Websocket, true)},
			expectedIDs:   []channel.ID{channelIDs[4]},
			expectedCount: 1,
		},
		{
			id: "7",
			options: channel.ReadOptions{
				UserIDEquals: c.NewOptional(s.user.ID, true),
				TypeEquals:   c.NewOptional(channel.Email, true),
			},
			expectedIDs:   []channel.ID{channelIDs[0], channelIDs[5]},
			expectedCount: 2,
		},
		{
			id: "8",
			options: channel.ReadOptions{
				UserIDEquals: c.NewOptional(s.otherUser.ID, true),
				TypeEquals:   c.NewOptional(channel.Telegram, true),
			},
			expectedIDs:   []channel.ID{channelIDs[3], channelIDs[6]},
			expectedCount: 2,
		},
		{
			id: "9",
			options: channel.ReadOptions{
				UserIDEquals: c.NewOptional(s.user.ID, true),
				TypeEquals:   c.NewOptional(channel.Websocket, true),
			},
			expectedIDs:   []channel.ID{},
			expectedCount: 0,
		},
		{
			id: "10",
			options: channel.ReadOptions{
				IDIn: c.NewOptional([]channel.ID{}, true),
			},
			expectedIDs:   []channel.ID{},
			expectedCount: 0,
		},
		{
			id: "11",
			options: channel.ReadOptions{
				IDIn: c.NewOptional([]channel.ID{channelIDs[0]}, true),
			},
			expectedIDs:   []channel.ID{channelIDs[0]},
			expectedCount: 1,
		},
		{
			id: "12",
			options: channel.ReadOptions{
				IDIn: c.NewOptional([]channel.ID{channelIDs[0], channelIDs[6], channelIDs[3]}, true),
			},
			expectedIDs:   []channel.ID{channelIDs[0], channelIDs[3], channelIDs[6]},
			expectedCount: 3,
		},
		{
			id: "13",
			options: channel.ReadOptions{
				IDIn: c.NewOptional(
					[]channel.ID{channelIDs[0], channelIDs[1], channelIDs[2], channelIDs[3]},
					true,
				),
				UserIDEquals: c.NewOptional(s.otherUser.ID, true),
				TypeEquals:   c.NewOptional(channel.Telegram, true),
			},
			expectedIDs:   []channel.ID{channelIDs[3]},
			expectedCount: 1,
		},
		{
			id: "14",
			options: channel.ReadOptions{
				UserIDEquals: c.NewOptional(s.user.ID, true),
				TypeEquals:   c.NewOptional(channel.Email, true),
				OrderBy:      channel.OrderByIDDesc,
			},
			expectedIDs:   []channel.ID{channelIDs[5], channelIDs[0]},
			expectedCount: 2,
		},
		{
			id: "15",
			options: channel.ReadOptions{
				TypeEquals: c.NewOptional(channel.Telegram, true),
				Limit:      c.NewOptional(uint(2), true),
			},
			expectedIDs:   []channel.ID{channelIDs[2], channelIDs[3]},
			expectedCount: 3,
		},
		{
			id: "16",
			options: channel.ReadOptions{
				OrderBy: channel.OrderByIDDesc,
				Limit:   c.NewOptional(uint(1), true),
			},
			expectedIDs:   []channel.ID{channelIDs[6]},
			expectedCount: 7,
		},
		{
			id: "17",
			options: channel.ReadOptions{
				UserIDEquals:    c.NewOptional(s.user.ID, true),
				IsDefaultEquals: c.NewOptional(true, true),
				Limit:           c.NewOptional(uint(1), true),
			},
			expectedIDs:   []channel.ID{channelIDs[0]},
			expectedCount: 1,
		},
		{
			id: "18",
			options: channel.ReadOptions{
				UserIDEquals:    c.NewOptional(s.otherUser.ID, true),
				IsDefaultEquals: c.NewOptional(true, true),
				Limit:           c.NewOptional(uint(1), true),
			},
			expectedIDs:   []channel.ID{channelIDs[3]},
			expectedCount: 1,
		},
		{
			id: "19",
			options: channel.ReadOptions{
				IsDefaultEquals: c.NewOptional(true, true),
				OrderBy:         channel.OrderByIDDesc,
			},
			expectedIDs:   []channel.ID{channelIDs[3], channelIDs[0]},
			expectedCount: 2,
		},
	}
	for _, testcase := range cases {
		actualIDs := s.readChannelIDs(testcase.options)
		actualCount := s.getCount(testcase.options)

		s.Equal(testcase.expectedIDs, actualIDs, testcase.id)
		s.Equal(testcase.expectedCount, actualCount, testcase.id)
	}
}

func (s *testSuite) createChannel(t channel.Type, u user.User, isDefault bool) channel.ID {
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
			CreatedAt:         Now,
			VerificationToken: c.NewOptional(channel.VerificationToken("test-2"), true),
			IsDefault:         isDefault,
		},
	)
	s.Nil(err)
	return createdChannel.ID
}

func (s *testSuite) TestGetByID() {
	channelOneID := s.createChannel(channel.Email, s.user, true)
	channelTwoID := s.createChannel(channel.Email, s.otherUser, false)

	assert := s.Require()
	channelOne, err := s.repo.GetByID(context.Background(), channelOneID)
	assert.Nil(err)
	assert.Equal(channelOneID, channelOne.ID)
	assert.Equal(channel.Email, channelOne.Type)
	assert.True(channelOne.IsDefault)
	assert.Equal(s.user.ID, channelOne.CreatedBy)
	assert.Equal(Now, channelOne.CreatedAt)

	channelTwo, err := s.repo.GetByID(context.Background(), channelTwoID)
	assert.Nil(err)
	assert.Equal(channelTwoID, channelTwo.ID)
	assert.Equal(channel.Email, channelTwo.Type)
	assert.False(channelTwo.IsDefault)
	assert.Equal(s.otherUser.ID, channelTwo.CreatedBy)
	assert.Equal(Now, channelTwo.CreatedAt)

	_, err = s.repo.GetByID(context.Background(), channel.ID(123456))
	assert.ErrorIs(err, channel.ErrChannelDoesNotExist)
}

func (s *testSuite) TestUpdate() {
	cases := []struct {
		id                 string
		tokenBefore        c.Optional[channel.VerificationToken]
		verifiedAtBefore   c.Optional[time.Time]
		typeBefore         channel.Type
		settingsBefore     channel.Settings
		doTokenUpdate      bool
		token              c.Optional[channel.VerificationToken]
		doVerifiedAtUpdate bool
		verifiedAt         c.Optional[time.Time]
		doSettingsUpdate   bool
		settings           channel.Settings
		expectedToken      c.Optional[channel.VerificationToken]
		expectedVerifiedAt c.Optional[time.Time]
		expectedSettings   channel.Settings
	}{
		{
			id:               "1",
			typeBefore:       channel.Email,
			settingsBefore:   channel.NewEmailSettings(c.NewEmail("test@test.test")),
			tokenBefore:      c.NewOptional(channel.VerificationToken("test"), true),
			verifiedAtBefore: c.NewOptional(Now, true),

			expectedToken:      c.NewOptional(channel.VerificationToken("test"), true),
			expectedVerifiedAt: c.NewOptional(Now, true),
			expectedSettings:   channel.NewEmailSettings(c.NewEmail("test@test.test")),
		},
		{
			id:               "2",
			typeBefore:       channel.Email,
			settingsBefore:   channel.NewEmailSettings(c.NewEmail("test@test.test")),
			tokenBefore:      c.NewOptional(channel.VerificationToken("test"), true),
			verifiedAtBefore: c.NewOptional(Now, true),

			doTokenUpdate: true,
			token:         c.NewOptional(channel.VerificationToken("token-after-update"), true),

			expectedToken:      c.NewOptional(channel.VerificationToken("token-after-update"), true),
			expectedVerifiedAt: c.NewOptional(Now, true),
			expectedSettings:   channel.NewEmailSettings(c.NewEmail("test@test.test")),
		},
		{
			id:             "3",
			typeBefore:     channel.Email,
			settingsBefore: channel.NewEmailSettings(c.NewEmail("test@test.test")),
			tokenBefore:    c.NewOptional(channel.VerificationToken("test"), true),

			doVerifiedAtUpdate: true,
			verifiedAt:         c.NewOptional(Now, true),

			expectedToken:      c.NewOptional(channel.VerificationToken("test"), true),
			expectedVerifiedAt: c.NewOptional(Now, true),
			expectedSettings:   channel.NewEmailSettings(c.NewEmail("test@test.test")),
		},
		{
			id:             "4",
			typeBefore:     channel.Email,
			settingsBefore: channel.NewEmailSettings(c.NewEmail("test@test.test")),
			tokenBefore:    c.NewOptional(channel.VerificationToken("test"), true),

			doTokenUpdate:      true,
			token:              c.NewOptional(channel.VerificationToken(""), false),
			doVerifiedAtUpdate: true,
			verifiedAt:         c.NewOptional(OtherTime, true),

			expectedToken:      c.NewOptional(channel.VerificationToken(""), false),
			expectedVerifiedAt: c.NewOptional(OtherTime, true),
			expectedSettings:   channel.NewEmailSettings(c.NewEmail("test@test.test")),
		},
		{
			id:             "5",
			typeBefore:     channel.Telegram,
			settingsBefore: channel.NewTelegramSettings(channel.TelegramBot("test"), channel.TelegramChatID(0)),
			tokenBefore:    c.NewOptional(channel.VerificationToken("test"), true),

			doTokenUpdate:      true,
			token:              c.NewOptional(channel.VerificationToken(""), false),
			doVerifiedAtUpdate: true,
			verifiedAt:         c.NewOptional(OtherTime, true),
			doSettingsUpdate:   true,
			settings:           channel.NewTelegramSettings(channel.TelegramBot("test"), channel.TelegramChatID(123)),

			expectedToken:      c.NewOptional(channel.VerificationToken(""), false),
			expectedVerifiedAt: c.NewOptional(OtherTime, true),
			expectedSettings:   channel.NewTelegramSettings(channel.TelegramBot("test"), channel.TelegramChatID(123)),
		},
		{
			id:             "6",
			typeBefore:     channel.Telegram,
			settingsBefore: channel.NewTelegramSettings(channel.TelegramBot("test"), channel.TelegramChatID(0)),
			tokenBefore:    c.NewOptional(channel.VerificationToken("test"), true),

			doSettingsUpdate: true,
			settings:         channel.NewTelegramSettings(channel.TelegramBot("test"), channel.TelegramChatID(1)),

			expectedToken:    c.NewOptional(channel.VerificationToken("test"), true),
			expectedSettings: channel.NewTelegramSettings(channel.TelegramBot("test"), channel.TelegramChatID(1)),
		},
	}

	for _, testcase := range cases {
		createdChannel, err := s.repo.Create(context.Background(), channel.CreateInput{
			CreatedBy:         s.user.ID,
			Type:              testcase.typeBefore,
			Settings:          testcase.settingsBefore,
			CreatedAt:         Now,
			VerificationToken: testcase.tokenBefore,
			VerifiedAt:        testcase.verifiedAtBefore,
		})
		assert := s.Require()
		assert.Nil(err, testcase.id)

		updateChannel, err := s.repo.Update(context.Background(), channel.UpdateInput{
			ID:                        createdChannel.ID,
			DoVerificationTokenUpdate: testcase.doTokenUpdate,
			VerificationToken:         testcase.token,
			DoVerifiedAtUpdate:        testcase.doVerifiedAtUpdate,
			VerifiedAt:                testcase.verifiedAt,
			DoSettingsUpdate:          testcase.doSettingsUpdate,
			Settings:                  testcase.settings,
		})
		assert.Nil(err, testcase.id)
		assert.Equal(createdChannel.ID, updateChannel.ID, testcase.id)
		assert.Equal(createdChannel.Type, updateChannel.Type, testcase.id)
		assert.Equal(createdChannel.CreatedAt, updateChannel.CreatedAt, testcase.id)
		assert.Equal(createdChannel.CreatedBy, updateChannel.CreatedBy, testcase.id)
		assert.Equal(testcase.expectedSettings, updateChannel.Settings, testcase.id)
		assert.Equal(testcase.expectedToken, updateChannel.VerificationToken, testcase.id)
		assert.Equal(testcase.expectedVerifiedAt, updateChannel.VerifiedAt, testcase.id)
	}
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
