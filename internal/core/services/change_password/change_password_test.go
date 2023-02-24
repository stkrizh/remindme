package changepassword

import (
	"context"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"testing"

	"github.com/stretchr/testify/require"
)

const USER_ID = 123

type suite struct {
	log      *logging.FakeLogger
	userRepo *user.FakeUserRepository
	hasher   *user.FakePasswordHasher
}

func setupSuite() *suite {
	userRepo := user.NewFakeUserRepository()
	userRepo.Users = []user.User{{ID: USER_ID}}
	return &suite{
		log:      logging.NewFakeLogger(),
		userRepo: userRepo,
		hasher:   user.NewFakePasswordHasher(),
	}
}

func (s *suite) createService() services.Service[Input, Result] {
	return New(s.log, s.userRepo, s.hasher)
}

func TestPasswordSuccessfullyChanged(t *testing.T) {
	cases := []struct {
		id                      string
		userID                  user.ID
		currentPassswordInDB    string
		currentPassswordInInput string
		newPasswordInInput      string
	}{
		{
			id:                      "1",
			userID:                  USER_ID,
			currentPassswordInDB:    "test-1",
			currentPassswordInInput: "test-1",
			newPasswordInInput:      "test-2",
		},
		{
			id:                      "2",
			userID:                  USER_ID,
			currentPassswordInDB:    "test-2",
			currentPassswordInInput: "test-2",
			newPasswordInInput:      "test-2",
		},
		{
			id:                      "3",
			userID:                  USER_ID,
			currentPassswordInDB:    "aaa",
			currentPassswordInInput: "aaa",
			newPasswordInInput:      "bbb",
		},
	}

	for _, testcase := range cases {
		t.Run(testcase.id, func(t *testing.T) {
			// Setup ---
			suite := setupSuite()
			service := suite.createService()

			// Exercise ---
			input := Input{
				CurrentPassword: user.RawPassword(testcase.currentPassswordInInput),
				NewPassword:     user.RawPassword(testcase.newPasswordInInput),
			}
			input.User.ID = testcase.userID
			input.User.PasswordHash = c.NewOptional(
				hashPassword(testcase.currentPassswordInDB, suite.hasher),
				true,
			)
			_, err := service.Run(context.Background(), input)

			// Verify ---
			require.NoError(t, err)
			assertPasswordValid(t, suite, testcase.newPasswordInInput)
		})
	}
}

func TestCurrentPasswordInvalid(t *testing.T) {
	// Setup ---
	suite := setupSuite()
	service := suite.createService()

	// Exercise ---
	input := Input{
		CurrentPassword: user.RawPassword("invalid-password"),
		NewPassword:     user.RawPassword("bbb"),
	}
	input.User.ID = USER_ID
	input.User.PasswordHash = c.NewOptional(
		hashPassword("valid-password", suite.hasher),
		true,
	)
	_, err := service.Run(context.Background(), input)

	// Verify ---
	require.ErrorIs(t, err, user.ErrInvalidCredentials)
}

func hashPassword(raw string, hasher user.PasswordHasher) user.PasswordHash {
	hash, err := hasher.HashPassword(user.RawPassword(raw))
	if err != nil {
		panic(err)
	}
	return hash
}

func assertPasswordValid(t *testing.T, suite *suite, password string) {
	t.Helper()

	u, err := suite.userRepo.GetByID(context.Background(), USER_ID)
	require.NoError(t, err)

	isValid := suite.hasher.ValidatePassword(user.RawPassword(password), u.PasswordHash.Value)
	require.True(t, isValid)
}
