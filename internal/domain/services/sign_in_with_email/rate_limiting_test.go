package signinwithemail

import (
	"context"
	"remindme/internal/domain/logging"
	ratelimiter "remindme/internal/domain/rate_limiter"
	"remindme/internal/domain/user"
	"testing"

	"github.com/stretchr/testify/suite"
)

type stubSignInWithEmailService struct {
	WasCalled bool
}

func NewStubSignUpAnonymouslyService() *stubSignInWithEmailService {
	return &stubSignInWithEmailService{}
}

func (s *stubSignInWithEmailService) Run(ctx context.Context, input Input) (result Result, err error) {
	s.WasCalled = true
	return result, nil
}

type testRateLimitingSuite struct {
	suite.Suite
	Logger      *logging.FakeLogger
	RateLimiter *ratelimiter.FakeRateLimiter
	Inner       *stubSignInWithEmailService
	Service     *serviceWithRateLimiting
}

func (suite *testRateLimitingSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.RateLimiter = ratelimiter.NewFakeRateLimiter(false)
	suite.Inner = NewStubSignUpAnonymouslyService()
	suite.Service = NewWithRateLimiting(
		suite.Logger,
		suite.RateLimiter,
		ratelimiter.Limit{Value: 10, Interval: ratelimiter.Minute},
		suite.Inner,
	)
}

func TestRateLimitingSuite(t *testing.T) {
	suite.Run(t, new(testRateLimitingSuite))
}

func (suite *testRateLimitingSuite) TestNotLimited() {
	ctx := context.Background()
	suite.RateLimiter.IsAllowed = true
	_, err := suite.Service.Run(ctx, Input{Email: user.NewEmail("test@test.test"), Password: user.RawPassword("test")})

	assert := suite.Require()
	assert.Nil(err)
	assert.True(suite.Inner.WasCalled)
}

func (suite *testRateLimitingSuite) TestLimited() {
	ctx := context.Background()
	suite.RateLimiter.IsAllowed = false
	suite.Service.Run(ctx, Input{Email: user.NewEmail("test@test.test"), Password: user.RawPassword("test")})

	assert := suite.Require()
	assert.False(suite.Inner.WasCalled)
}
