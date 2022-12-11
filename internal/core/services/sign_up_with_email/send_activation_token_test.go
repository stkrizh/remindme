package signupwithemail

import (
	"context"
	"errors"
	"fmt"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"testing"

	"github.com/stretchr/testify/suite"
)

const PASSWORD = user.RawPassword("test")

var errTest = fmt.Errorf("test error")

type stubSignUpService struct {
	err error
}

func newStubSignUpService(err error) *stubSignUpService {
	return &stubSignUpService{err: err}
}

func (s *stubSignUpService) Run(ctx context.Context, input Input) (result Result, err error) {
	return result, s.err
}

type testActivationSuite struct {
	suite.Suite
	Logger  *logging.FakeLogger
	Sender  *user.FakeActivationTokenSender
	Inner   *stubSignUpService
	Service *serviceWithActivationTokenSending
}

func (suite *testActivationSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.Sender = user.NewFakeActivationTokenSender()
	suite.Inner = newStubSignUpService(nil)
	suite.Service = NewWithActivationTokenSending(
		suite.Logger,
		suite.Sender,
		suite.Inner,
	)
}

func TestSendActivationTokenService(t *testing.T) {
	suite.Run(t, new(testActivationSuite))
}

func (suite *testActivationSuite) TestActivationEmailSent() {
	ctx := context.Background()
	_, err := suite.Service.Run(ctx, Input{Email: EMAIL, Password: PASSWORD})

	assert := suite.Require()
	assert.Nil(err)
	assert.Equal(1, suite.Sender.SentCount())
}

func (suite *testActivationSuite) TestSignUpServiceError() {
	service := NewWithActivationTokenSending(
		suite.Logger,
		suite.Sender,
		newStubSignUpService(errTest),
	)
	ctx := context.Background()
	_, err := service.Run(ctx, Input{Email: EMAIL, Password: PASSWORD})

	assert := suite.Require()
	assert.NotNil(err)
	assert.True(errors.Is(err, errTest))
	assert.Equal(0, suite.Sender.SentCount())
}
