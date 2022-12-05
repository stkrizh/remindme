package ratelimiter

import "context"

type Interval int

const (
	Minute = Interval(0)
	Hour   = Interval(1)
)

type Limit struct {
	Value    uint16
	Interval Interval
}

type RateLimiter interface {
	CheckLimit(ctx context.Context, key string, limit Limit) error
}
