package ratelimiter

import "context"

type FakeRateLimiter struct {
	IsAllowed bool
}

func NewFakeRateLimiter(isAllowed bool) *FakeRateLimiter {
	return &FakeRateLimiter{IsAllowed: isAllowed}
}

func (rl *FakeRateLimiter) CheckLimit(ctx context.Context, key string, limit Limit) Result {
	if rl.IsAllowed {
		return Result{IsAllowed: true}
	}
	return Result{IsAllowed: false}
}
