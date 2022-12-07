package signupanonymously

import (
	"context"
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
	rate := s.rateLimiter.CheckLimit(ctx, rateLimitKey, s.rateLimit)
	if rate.IsAllowed {
		return s.inner.Run(ctx, input)
	}

	s.log.Warning(
		ctx,
		"Rate limit exceeded for anonymous signing up.",
		logging.Entry("ip", input.IP),
	)
	return result, ratelimiter.ErrRateLimitExceeded
}
