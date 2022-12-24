package ratelimiting

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	ratelimiter "remindme/internal/core/domain/rate_limiter"
	"remindme/internal/core/services"
)

type hasRateLimitKey interface {
	GetRateLimitKey() string
}

type serviceWithRateLimiting[T hasRateLimitKey, S any] struct {
	log         logging.Logger
	rateLimiter ratelimiter.RateLimiter
	rateLimit   ratelimiter.Limit
	inner       services.Service[T, S]
}

func WithRateLimiting[T hasRateLimitKey, S any](
	log logging.Logger,
	rateLimiter ratelimiter.RateLimiter,
	rateLimit ratelimiter.Limit,
	inner services.Service[T, S],
) services.Service[T, S] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if rateLimiter == nil {
		panic(e.NewNilArgumentError("rateLimiter"))
	}
	return &serviceWithRateLimiting[T, S]{
		log:         log,
		rateLimiter: rateLimiter,
		rateLimit:   rateLimit,
		inner:       inner,
	}
}

func (s *serviceWithRateLimiting[T, S]) Run(ctx context.Context, input T) (result S, err error) {
	rateLimitKey := input.GetRateLimitKey()
	rate := s.rateLimiter.CheckLimit(ctx, rateLimitKey, s.rateLimit)
	if rate.IsAllowed {
		return s.inner.Run(ctx, input)
	}

	s.log.Warning(ctx, "Rate limit exceeded.", logging.Entry("key", rateLimitKey))
	return result, ratelimiter.ErrRateLimitExceeded
}
