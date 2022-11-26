package signupwithemail

import (
	"context"
	"remindme/internal/domain/common"
	"remindme/internal/domain/logging"
	uow "remindme/internal/domain/unit_of_work"
	"remindme/internal/domain/user"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const ACTIVATION_TOKEN = "test"

func TestSignUpWithEmailService(t *testing.T) {
	logger := logging.NewFakeLogger()
	userRepository := user.NewFakeRepository()
	unitOfWorkContext := uow.NewFakeUnitOfWorkContext(userRepository)
	unitOfWork := uow.NewFakeUnitOfWork(unitOfWorkContext)
	passwordHasher := user.NewFakePasswordHasher()
	activationTokenGenerator := user.NewFakeActivationTokenGenerator(ACTIVATION_TOKEN)
	activationTokenSender := user.NewFakeActivationTokenSender()
	now := time.Now().UTC()

	service := New(
		logger,
		unitOfWork,
		passwordHasher,
		activationTokenGenerator,
		activationTokenSender,
		func() time.Time { return now },
	)

	context := context.Background()
	result, err := service.Run(context, Input{Email: user.Email("test@test.com"), Password: user.RawPassword("test")})

	assert := require.New(t)
	assert.Nil(err)
	assert.NotNil(result)

	assert.Equal(1, activationTokenSender.SentCount())
	u := activationTokenSender.LastSentTo()
	assert.Equal(common.NewOptional(user.Email("test@test.com"), true), u.Email)
	assert.Equal(common.NewOptional(user.ActivationToken(ACTIVATION_TOKEN), true), u.ActivationToken)
}
