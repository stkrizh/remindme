package ratelimiter

import "context"

type FakeRateLimiter struct {
	ReturnRateLimitError bool
}

func NewFakeRateLimiter(returnRateLimitError bool) *FakeRateLimiter {
	return &FakeRateLimiter{ReturnRateLimitError: returnRateLimitError}
}

func (rl *FakeRateLimiter) CheckLimit(ctx context.Context, key string, limit Limit) error {
	if rl.ReturnRateLimitError {
		return NewRateLimiteExceededError(key)
	}
	return nil
}
