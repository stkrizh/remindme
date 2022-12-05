package signupwithemail

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

const PASSWORD = user.RawPassword("test")

var errTest = fmt.Errorf("test error")

type stubSignUpService struct {
	err error
}

func newStubSignUpService(err error) services.Service[Input, Result] {
	return &stubSignUpService{err: err}
}

func (s *stubSignUpService) Run(ctx context.Context, input Input) (result Result, err error) {
	if s.err != nil {
		return result, s.err
	}
	return result, s.err
}

func TestActivationEmailSent(t *testing.T) {
	logger := logging.NewFakeLogger()
	sender := user.NewFakeActivationTokenSender()
	stubSignUpService := newStubSignUpService(nil)
	service := NewWithActivationTokenSending(logger, sender, stubSignUpService)

	ctx := context.Background()
	_, err := service.Run(ctx, Input{Email: EMAIL, Password: PASSWORD})

	assert := require.New(t)
	assert.Nil(err)
	assert.Equal(1, sender.SentCount())
}

func TestSignUpServiceError(t *testing.T) {
	logger := logging.NewFakeLogger()
	sender := user.NewFakeActivationTokenSender()
	stubSignUpService := newStubSignUpService(errTest)
	service := NewWithActivationTokenSending(logger, sender, stubSignUpService)

	ctx := context.Background()
	_, err := service.Run(ctx, Input{Email: EMAIL, Password: PASSWORD})

	assert := require.New(t)
	assert.NotNil(err)
	assert.True(errors.Is(err, errTest))
}
