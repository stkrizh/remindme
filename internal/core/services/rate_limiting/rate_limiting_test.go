package ratelimiting

import (
	"context"
	"remindme/internal/core/domain/logging"
	ratelimiter "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/services"
	"testing"

	"github.com/stretchr/testify/suite"
)

type input struct {
	Value string
}

func (i input) GetRateLimitKey() string {
	return "test-rate-limiting-key::" + i.Value
}

type result struct{}

type stubService struct {
	WasCalled bool
}

func NewStubService() services.Service[input, result] {
	return &stubService{}
}

func (s *stubService) Run(ctx context.Context, input input) (result result, err error) {
	s.WasCalled = true
	return result, nil
}

type testRateLimitingSuite struct {
	suite.Suite
	Logger      *logging.FakeLogger
	RateLimiter *ratelimiter.FakeRateLimiter
	Inner       services.Service[input, result]
	Service     services.Service[input, result]
}

func (suite *testRateLimitingSuite) SetupTest() {
	suite.Logger = logging.NewFakeLogger()
	suite.RateLimiter = ratelimiter.NewFakeRateLimiter(false)
	suite.Inner = NewStubService()
	suite.Service = New(
		suite.Logger,
		suite.RateLimiter,
		ratelimiter.Limit{Value: 10, Interval: ratelimiter.Minute},
		suite.Inner,
	)
}

func TestRateLimitingService(t *testing.T) {
	suite.Run(t, new(testRateLimitingSuite))
}

func (suite *testRateLimitingSuite) TestNotLimited() {
	ctx := context.Background()
	suite.RateLimiter.IsAllowed = true
	_, err := suite.Service.Run(ctx, input{Value: "test"})

	assert := suite.Require()
	assert.Nil(err)
	innerService, ok := suite.Inner.(*stubService)
	assert.True(ok)
	assert.True(innerService.WasCalled)
}

func (suite *testRateLimitingSuite) TestLimited() {
	ctx := context.Background()
	suite.RateLimiter.IsAllowed = false
	suite.Service.Run(ctx, input{Value: "test"})

	assert := suite.Require()
	innerService, ok := suite.Inner.(*stubService)
	assert.True(ok)
	assert.False(innerService.WasCalled)
}
