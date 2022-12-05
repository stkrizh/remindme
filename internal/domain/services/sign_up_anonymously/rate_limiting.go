package signupanonymously

import (
	"context"
	"errors"
	e "remindme/internal/domain/errors"
	"remindme/internal/domain/logging"
	ratelimiter "remindme/internal/domain/rate_limiter"
	"remindme/internal/domain/services"
)

type serviceWithRateLimiting struct {
	log         logging.Logger
	rateLimiter ratelimiter.RateLimiter
	rateLimit   ratelimiter.Limit
	inner       services.Service[Input, Result]
}

func NewWithRateLimiting(
	log logging.Logger,
	rateLimiter ratelimiter.RateLimiter,
	rateLimit ratelimiter.Limit,
	inner services.Service[Input, Result],
) *serviceWithRateLimiting {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if rateLimiter == nil {
		panic(e.NewNilArgumentError("rateLimiter"))
	}
	return &serviceWithRateLimiting{
		log:         log,
		rateLimiter: rateLimiter,
		rateLimit:   rateLimit,
		inner:       inner,
	}
}

func (s *serviceWithRateLimiting) Run(ctx context.Context, input Input) (result Result, err error) {
	rateLimitKey := "sign-up-anonymously::" + input.IP.String()
	err = s.rateLimiter.CheckLimit(ctx, rateLimitKey, s.rateLimit)
	if err == nil {
		return s.inner.Run(ctx, input)
	}

	if errors.Is(err, context.Canceled) {
		return result, err
	}
	var errRateLimitExceed *ratelimiter.RateLimitExceededError
	if errors.As(err, &errRateLimitExceed) {
		s.log.Warning(
			ctx,
			"Rate limit exceeded for anonymous signing up.",
			logging.Entry("ip", input.IP),
		)
		return result, err
	}
	s.log.Error(
		ctx,
		"Could not check rate limit for anonymous signing up.",
		logging.Entry("input", input),
		logging.Entry("err", err),
	)
	return result, err
}
