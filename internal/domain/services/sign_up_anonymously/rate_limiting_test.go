package signupanonymously

import (
	"context"
	"net/netip"
	"remindme/internal/domain/logging"
	ratelimiter "remindme/internal/domain/rate_limiter"
	"testing"

	"github.com/stretchr/testify/suite"
)

type stubSignUpAnonymouslyService struct {
	WasCalled bool
}

func NewStubSignUpAnonymouslyService() *stubSignUpAnonymouslyService {
	return &stubSignUpAnonymouslyService{}
}

func (s *stubSignUpAnonymouslyService) Run(ctx context.Context, input Input) (result Result, err error) {
	s.WasCalled = true
	return result, nil
}

type testRateLimitingSuite struct {
	suite.Suite
	Logger      *logging.FakeLogger
	RateLimiter *ratelimiter.FakeRateLimiter
	Inner       *stubSignUpAnonymouslyService
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
	_, err := suite.Service.Run(ctx, Input{IP: netip.MustParseAddr("192.168.1.1")})

	assert := suite.Require()
	assert.Nil(err)
	assert.True(suite.Inner.WasCalled)
}

func (suite *testRateLimitingSuite) TestLimited() {
	ctx := context.Background()
	suite.RateLimiter.IsAllowed = false
	suite.Service.Run(ctx, Input{IP: netip.MustParseAddr("192.168.1.1")})

	assert := suite.Require()
	assert.False(suite.Inner.WasCalled)
}
