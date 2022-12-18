package channel

import (
	"context"
	"reflect"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/common"
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
		settings channel.Settings
	}{
		{id: "email-1", settings: &channel.EmailSettings{Email: common.NewEmail("test-1@test.com")}},
		{id: "email-2", settings: &channel.EmailSettings{Email: common.NewEmail("test-2@test.com")}},
		{id: "telegram-1", settings: &channel.TelegramSettings{
			BotToken: channel.TelegramBotToken("test"),
			ChatID:   channel.TelegramChatID(1),
		}},
		{id: "telegram-2", settings: &channel.TelegramSettings{
			BotToken: channel.TelegramBotToken("test-test-test-test aasdkhaskdhsakdhjksahdkjashd jkahsdkjas"),
			ChatID:   channel.TelegramChatID(111222333444555),
		}},
		{id: "ws-1", settings: &channel.WebsocketSettings{}},
	}
	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			encoded, err := encodeSettings(testcase.settings)
			if err != nil {
				t.Fatal("could not encode channel settings:", err)
			}

			decoded, err := decodeSettings(encoded)
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
	pool     *pgxpool.Pool
	repo     *PgxChannelRepository
	userRepo *dbuser.PgxUserRepository
	user     user.User
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
			Email:        common.NewOptional(common.NewEmail("test@test.test"), true),
			PasswordHash: common.NewOptional(user.PasswordHash("test"), true),
			CreatedAt:    NOW,
			ActivatedAt:  common.NewOptional(NOW, true),
		},
	)
	s.Nil(err)
	s.user = u
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
				CreatedBy:  s.user.ID,
				Settings:   channel.NewEmailSettings("test-1@test.test"),
				CreatedAt:  time.Now().UTC(),
				IsVerified: true,
			},
		},
		{
			id: "email-2",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Settings:   channel.NewEmailSettings("test-2@test.test"),
				CreatedAt:  NOW,
				IsVerified: false,
			},
		},
		{
			id: "telegram-1",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Settings:   channel.NewTelegramSettings("test-1", 1),
				CreatedAt:  time.Now().UTC(),
				IsVerified: true,
			},
		},
		{
			id: "telegram-2",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Settings:   channel.NewTelegramSettings("test-2", 111222333444),
				CreatedAt:  NOW,
				IsVerified: false,
			},
		},
		{
			id: "websocket-1",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Settings:   channel.NewWebsocketSettings(),
				CreatedAt:  time.Now().UTC(),
				IsVerified: true,
			},
		},
		{
			id: "websocket-2",
			input: channel.CreateInput{
				CreatedBy:  s.user.ID,
				Settings:   channel.NewWebsocketSettings(),
				CreatedAt:  NOW,
				IsVerified: false,
			},
		},
	}

	for _, testcase := range cases {
		s.Run(testcase.id, func() {
			newChannel, err := s.repo.Create(context.Background(), testcase.input)

			s.Nil(err)
			s.Equal(testcase.input.CreatedBy, newChannel.CreatedBy)
			s.Equal(testcase.input.Settings, newChannel.Settings)
			s.Equal(
				testcase.input.CreatedAt.Truncate(time.Millisecond),
				newChannel.CreatedAt.Truncate(time.Millisecond),
			)
			s.Equal(testcase.input.IsVerified, newChannel.IsVerified)
		})
	}

}
