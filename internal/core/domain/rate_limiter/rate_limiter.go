package ratelimiter

import (
	"context"
	"errors"
)

var ErrRateLimitExceeded = errors.New("rate limit exceeded")

type Interval struct {
	value int
}

var (
	Minute = Interval{}
	Hour   = Interval{value: 1}
)

type Limit struct {
	Value    uint16
	Interval Interval
}

type Result struct {
	IsAllowed bool
}

func Allowed() Result {
	return Result{IsAllowed: true}
}

func NotAllowed() Result {
	return Result{IsAllowed: false}
}

type RateLimiter interface {
	CheckLimit(ctx context.Context, key string, limit Limit) Result
}
