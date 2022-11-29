package sendactivationemail

import (
	"context"
	"errors"
	"fmt"
	"remindme/internal/domain/logging"
	"remindme/internal/domain/services"
	"remindme/internal/domain/user"
	"testing"

	"github.com/stretchr/testify/require"
)

const EMAIL = user.Email("test@test.test")
const PASSWORD = user.RawPassword("test")

var testErr = fmt.Errorf("test error")

type stubSignUpService struct {
	err error
}

func newStubSignUpService(err error) services.Service[services.SignUpWithEmailInput, services.SignUpWithEmailResult] {
	return &stubSignUpService{err: err}
}

func (s *stubSignUpService) Run(
	ctx context.Context,
	input services.SignUpWithEmailInput,
) (result services.SignUpWithEmailResult, err error) {
	if s.err != nil {
		return result, s.err
	}
	return result, s.err
}

func TestActivationEmailSent(t *testing.T) {
	logger := logging.NewFakeLogger()
	sender := user.NewFakeActivationTokenSender()
	stubSignUpService := newStubSignUpService(nil)
	service := New(logger, sender, stubSignUpService)

	ctx := context.Background()
	_, err := service.Run(ctx, services.SignUpWithEmailInput{Email: EMAIL, Password: PASSWORD})

	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(1, sender.SentCount())
}

func TestSignUpServiceError(t *testing.T) {
	logger := logging.NewFakeLogger()
	sender := user.NewFakeActivationTokenSender()
	stubSignUpService := newStubSignUpService(testErr)
	service := New(logger, sender, stubSignUpService)

	ctx := context.Background()
	_, err := service.Run(ctx, services.SignUpWithEmailInput{Email: EMAIL, Password: PASSWORD})

	assert := require.New(t)
	assert.NotNil(err)
	assert.True(errors.Is(err, testErr))
}

func TestActivationSendingError(t *testing.T) {
	logger := logging.NewFakeLogger()
	sender := user.NewFakeActivationTokenSender()
	sender.ReturnError = true
	stubSignUpService := newStubSignUpService(nil)
	service := New(logger, sender, stubSignUpService)

	ctx := context.Background()
	_, err := service.Run(ctx, services.SignUpWithEmailInput{Email: EMAIL, Password: PASSWORD})

	assert := require.New(t)
	assert.NotNil(err)
	var expectedErr *user.ActivationTokenSendingError
	assert.True(errors.As(err, &expectedErr))
}
