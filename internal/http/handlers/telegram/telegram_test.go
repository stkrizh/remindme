package telegram

import (
	"remindme/internal/core/domain/channel"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseVerificationDataSuccess(t *testing.T) {
	cases := []struct {
		id       string
		userID   int64
		text     string
		expected channelVerificationData
	}{
		{
			id:     "1",
			userID: 111222333,
			text:   "vrf-1-aaa",
			expected: channelVerificationData{
				channelID:      channel.ID(1),
				telegramChatID: channel.TelegramChatID(111222333),
				token:          channel.VerificationToken("aaa"),
			},
		},
		{
			id:     "2",
			userID: 1,
			text:   "vrf-111222333-aaaBBBcccDDD",
			expected: channelVerificationData{
				channelID:      channel.ID(111222333),
				telegramChatID: channel.TelegramChatID(1),
				token:          channel.VerificationToken("aaaBBBcccDDD"),
			},
		},
		{
			id:     "3",
			userID: 1,
			text:   "vrf-123-",
			expected: channelVerificationData{
				channelID:      channel.ID(123),
				telegramChatID: channel.TelegramChatID(1),
				token:          channel.VerificationToken(""),
			},
		},
		{
			id:     "4",
			userID: 123456789,
			text:   "vrf-987654321-aaa-bbb-ccc-ddd-eee-fff",
			expected: channelVerificationData{
				channelID:      channel.ID(987654321),
				telegramChatID: channel.TelegramChatID(123456789),
				token:          channel.VerificationToken("aaa-bbb-ccc-ddd-eee-fff"),
			},
		},
		{
			id:     "5",
			userID: 123456789,
			text:   "/start vrf-111-token",
			expected: channelVerificationData{
				channelID:      channel.ID(111),
				telegramChatID: channel.TelegramChatID(123456789),
				token:          channel.VerificationToken("token"),
			},
		},
	}
	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			update := createUpdate(testcase.userID, testcase.text)
			data, ok := parseVerificationData(update)

			assert := require.New(t)
			assert.True(ok)
			assert.Equal(testcase.expected, data)
		})
	}
}

func TestParseVerificationDataFail(t *testing.T) {
	cases := []struct {
		id   string
		text string
	}{
		{id: "1", text: ""},
		{id: "2", text: "vrf"},
		{id: "3", text: "vrf-aaa-aaaBBBcccDDD"},
		{id: "4", text: "vrf-123"},
		{id: "5", text: "vrf--123-token"},
		{id: "6", text: "123-token"},
	}
	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			update := createUpdate(1, testcase.text)
			_, ok := parseVerificationData(update)

			assert := require.New(t)
			assert.False(ok)
		})
	}
}

func createUpdate(userID int64, text string) update {
	return update{
		ID: 1,
		Message: &message{
			ID:   2,
			From: user{ID: userID},
			Date: time.Now().Unix(),
			Text: text,
		},
	}
}
