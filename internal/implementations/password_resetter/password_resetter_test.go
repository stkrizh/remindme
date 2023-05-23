package passwordresetter

import (
	"encoding/base64"
	"fmt"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	EMAIL            = "test@test.test"
	PASSWORD_HASH    = "test-password-hash"
	ACTIVATION_TOKEN = "test-activation-token"
)

var NOW time.Time = time.Now().UTC()

type testSuite struct {
	suite.Suite
	users map[user.ID]user.User
}

func (suite *testSuite) SetupTest() {
	suite.users = make(map[user.ID]user.User)
	suite.users[user.ID(1)] = user.User{
		ID:           user.ID(1),
		Email:        c.NewOptional(c.Email("test-1@test.test"), true),
		PasswordHash: c.NewOptional(user.PasswordHash("test-hash-1"), true),
		CreatedAt:    NOW,
		ActivatedAt:  c.NewOptional(NOW, true),
	}
	suite.users[user.ID(1234)] = user.User{
		ID:           user.ID(1234),
		Email:        c.NewOptional(c.Email("test-1234@test.test"), true),
		PasswordHash: c.NewOptional(user.PasswordHash("test-hash-1234"), true),
		CreatedAt:    NOW,
		ActivatedAt:  c.NewOptional(NOW, true),
	}
	suite.users[user.ID(111222333)] = user.User{
		ID:           user.ID(111222333),
		Email:        c.NewOptional(c.Email("test-111222333@test.test"), true),
		PasswordHash: c.NewOptional(user.PasswordHash("test-hash-111222333"), true),
		CreatedAt:    NOW,
		ActivatedAt:  c.NewOptional(NOW, true),
	}
}

func TestHMACPasswordResetter(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestSuccessCases() {
	cases := []struct {
		ID               string
		SecretKeyToGen   string
		SecretKeyToCheck string
		GenTime          string
		CheckTime        string
		ValidDuration    time.Duration
	}{
		{
			ID:               "1",
			SecretKeyToGen:   "",
			SecretKeyToCheck: "",
			GenTime:          "2020-01-01T15:00:00Z",
			CheckTime:        "2020-01-02T14:59:59Z",
			ValidDuration:    time.Hour * 24,
		},
		{
			ID:               "2",
			SecretKeyToGen:   "test",
			SecretKeyToCheck: "test",
			GenTime:          "2020-01-01T15:00:00Z",
			CheckTime:        "2020-01-01T15:59:59Z",
			ValidDuration:    time.Hour,
		},
		{
			ID:               "3",
			SecretKeyToGen:   "test-test-test",
			SecretKeyToCheck: "test-test-test",
			GenTime:          "2020-01-01T15:00:00Z",
			CheckTime:        "2020-01-11T14:59:59Z",
			ValidDuration:    time.Hour * 240,
		},
	}

	for userID, u := range s.users {
		for _, testCase := range cases {
			s.Run(fmt.Sprintf("%d-%s", userID, testCase.ID), func() {
				genTime, err := time.Parse(time.RFC3339, testCase.GenTime)
				if err != nil {
					s.FailNow("GenTime is invalid")
				}
				checkTime, err := time.Parse(time.RFC3339, testCase.CheckTime)
				if err != nil {
					s.FailNow("CheckTime is invalid")
				}

				generator := NewHMAC(
					testCase.SecretKeyToGen,
					testCase.ValidDuration,
					func() time.Time { return genTime },
				)
				token := generator.GenerateToken(u)

				validator := NewHMAC(
					testCase.SecretKeyToCheck,
					testCase.ValidDuration,
					func() time.Time { return checkTime },
				)
				if !validator.ValidateToken(u, token) {
					s.FailNow("token validation failed", token)
				}
			})
		}
	}
}

func (s *testSuite) TestFailCases() {
	cases := []struct {
		ID               string
		SecretKeyToGen   string
		SecretKeyToCheck string
		GenTime          string
		CheckTime        string
		ValidDuration    time.Duration
	}{
		{
			ID:               "1",
			SecretKeyToGen:   "",
			SecretKeyToCheck: " ",
			GenTime:          "2020-01-01T15:00:00Z",
			CheckTime:        "2020-01-02T14:59:59Z",
			ValidDuration:    time.Hour * 24,
		},
		{
			ID:               "2",
			SecretKeyToGen:   "test",
			SecretKeyToCheck: " test",
			GenTime:          "2020-01-01T15:00:00Z",
			CheckTime:        "2020-01-01T15:59:59Z",
			ValidDuration:    time.Hour,
		},
		{
			ID:               "3",
			SecretKeyToGen:   "a",
			SecretKeyToCheck: "a",
			GenTime:          "2020-01-01T15:00:00Z",
			CheckTime:        "2020-01-02T15:00:01Z",
			ValidDuration:    time.Hour * 24,
		},
		{
			ID:               "4",
			SecretKeyToGen:   "test",
			SecretKeyToCheck: "test",
			GenTime:          "2020-01-01T15:00:00Z",
			CheckTime:        "2020-01-01T16:01:30Z",
			ValidDuration:    time.Hour,
		},
		{
			ID:               "5",
			SecretKeyToGen:   "test-test-test",
			SecretKeyToCheck: "test-test-test",
			GenTime:          "2020-01-01T15:00:00Z",
			CheckTime:        "2020-01-11T15:00:01Z",
			ValidDuration:    time.Hour * 240,
		},
	}

	for userID, u := range s.users {
		for _, testCase := range cases {
			s.Run(fmt.Sprintf("%d-%s", userID, testCase.ID), func() {
				genTime, err := time.Parse(time.RFC3339, testCase.GenTime)
				if err != nil {
					s.FailNow("GenTime is invalid")
				}
				checkTime, err := time.Parse(time.RFC3339, testCase.CheckTime)
				if err != nil {
					s.FailNow("CheckTime is invalid")
				}

				generator := NewHMAC(
					testCase.SecretKeyToGen,
					testCase.ValidDuration,
					func() time.Time { return genTime },
				)
				token := generator.GenerateToken(u)

				validator := NewHMAC(
					testCase.SecretKeyToCheck,
					testCase.ValidDuration,
					func() time.Time { return checkTime },
				)
				if validator.ValidateToken(u, token) {
					s.FailNow("token validation succeeded", token)
				}
			})
		}
	}
}

func (s *testSuite) TestFailForOtherUser() {
	resetter := NewHMAC(
		"test-secret-key",
		time.Hour*24,
		func() time.Time { return NOW },
	)
	token1 := resetter.GenerateToken(s.users[user.ID(1)])
	token1234 := resetter.GenerateToken(s.users[user.ID(1234)])
	s.False(resetter.ValidateToken(s.users[user.ID(1234)], token1))
	s.False(resetter.ValidateToken(s.users[user.ID(1)], token1234))
}

func (s *testSuite) TestFailIfUserIdModified() {
	resetter := NewHMAC(
		"test-secret-key",
		time.Hour*24,
		func() time.Time { return NOW },
	)
	u := s.users[user.ID(1)]
	token, err := base64.RawURLEncoding.DecodeString(string(resetter.GenerateToken(u)))
	s.Nil(err)

	u.ID = user.ID(2)
	parts := strings.SplitN(string(token), "-", 4)
	parts[0] = "2"
	invalidToken := user.PasswordResetToken(strings.Join(parts, "-"))

	s.False(resetter.ValidateToken(u, invalidToken))
}

func (s *testSuite) TestFailIfTimestampModified() {
	resetter := NewHMAC(
		"test-secret-key",
		time.Hour*24,
		func() time.Time { return NOW },
	)
	u := s.users[user.ID(1)]
	token, err := base64.RawURLEncoding.DecodeString(string(resetter.GenerateToken(u)))
	s.Nil(err)

	parts := strings.SplitN(string(token), "-", 4)
	ts, err := strconv.Atoi(parts[1])
	s.Nil(err)
	parts[1] = fmt.Sprintf("%d", ts-1)
	invalidToken := user.PasswordResetToken(strings.Join(parts, "-"))

	s.False(resetter.ValidateToken(u, invalidToken))
}

func (s *testSuite) TestFailIfSaltModified() {
	resetter := NewHMAC(
		"test-secret-key",
		time.Hour*24,
		func() time.Time { return NOW },
	)
	u := s.users[user.ID(1)]
	token, err := base64.RawURLEncoding.DecodeString(string(resetter.GenerateToken(u)))
	s.Nil(err)

	parts := strings.SplitN(string(token), "-", 4)
	parts[2] = " " + parts[2][1:]
	invalidToken := user.PasswordResetToken(strings.Join(parts, "-"))

	s.False(resetter.ValidateToken(u, invalidToken))
}

func (s *testSuite) TestGetUserID() {
	resetter := NewHMAC(
		"test-secret-key",
		time.Hour*24,
		func() time.Time { return NOW },
	)
	for userID, u := range s.users {
		s.Run(fmt.Sprintf("%d", userID), func() {
			token := resetter.GenerateToken(u)
			actualUserID, ok := resetter.GetUserID(token)
			s.True(ok)
			s.Equal(userID, actualUserID)
		})
	}
}
