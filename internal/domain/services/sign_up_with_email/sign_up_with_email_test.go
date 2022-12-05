package signupwithemail

import (
	"context"
	"errors"
	"remindme/internal/domain/common"
	"remindme/internal/domain/logging"
	uow "remindme/internal/domain/unit_of_work"
	"remindme/internal/domain/user"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const ACTIVATION_TOKEN = "test"
const EMAIL = user.Email("test@test.test")
const RAW_PASSWORD = user.RawPassword("test-password")

func TestSuccess(t *testing.T) {
	logger := logging.NewFakeLogger()
	userRepository := user.NewFakeUserRepository()
	sessionRepository := user.NewFakeSessionRepository(userRepository)
	unitOfWorkContext := uow.NewFakeUnitOfWorkContext(userRepository, sessionRepository)
	unitOfWork := uow.NewFakeUnitOfWork(unitOfWorkContext)
	passwordHasher := user.NewFakePasswordHasher()
	activationTokenGenerator := user.NewFakeActivationTokenGenerator(ACTIVATION_TOKEN)
	now := time.Now().UTC()

	service := New(
		logger,
		unitOfWork,
		passwordHasher,
		activationTokenGenerator,
		func() time.Time { return now },
	)

	context := context.Background()
	result, err := service.Run(context, Input{Email: EMAIL, Password: RAW_PASSWORD})

	assert := require.New(t)
	assert.Nil(err)
	assert.NotEqual(user.ID(0), result.User.ID)
	assert.Equal(now, result.User.CreatedAt)
	assert.True(result.User.Email.IsPresent)
	assert.Equal(EMAIL, result.User.Email.Value)
	assert.True(result.User.PasswordHash.IsPresent)
	assert.NotEqual(RAW_PASSWORD, result.User.PasswordHash.Value)
	assert.False(result.User.Identity.IsPresent)
	assert.True(unitOfWork.Context.WasCommitCalled)
}

func TestEmailAlreadyExistsError(t *testing.T) {
	logger := logging.NewFakeLogger()
	userRepository := user.NewFakeUserRepository()
	sessionRepository := user.NewFakeSessionRepository(userRepository)
	unitOfWorkContext := uow.NewFakeUnitOfWorkContext(userRepository, sessionRepository)
	unitOfWork := uow.NewFakeUnitOfWork(unitOfWorkContext)
	passwordHasher := user.NewFakePasswordHasher()
	activationTokenGenerator := user.NewFakeActivationTokenGenerator(ACTIVATION_TOKEN)
	now := time.Now().UTC()

	ctx := context.Background()
	userRepository.Create(
		ctx,
		user.CreateUserInput{
			Email:        common.NewOptional(EMAIL, true),
			PasswordHash: common.NewOptional(user.PasswordHash("test"), true),
			CreatedAt:    now,
		},
	)

	service := New(
		logger,
		unitOfWork,
		passwordHasher,
		activationTokenGenerator,
		func() time.Time { return now },
	)
	_, err := service.Run(ctx, Input{Email: EMAIL, Password: RAW_PASSWORD})

	assert := require.New(t)
	assert.NotNil(err)

	var expectedErr *user.EmailAlreadyExistsError
	assert.True(errors.As(err, &expectedErr))
	assert.False(unitOfWork.Context.WasCommitCalled)
	assert.True(unitOfWorkContext.WasRollbackCalled)
}
