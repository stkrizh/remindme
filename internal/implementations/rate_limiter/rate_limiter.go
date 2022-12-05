package ratelimiter

import (
	"context"
	"fmt"
	e "remindme/internal/domain/errors"
	ratelimiter "remindme/internal/domain/rate_limiter"
	"time"

	"github.com/go-redis/redis/v9"
)

type Redis struct {
	redisClient *redis.Client
	now         func() time.Time
}

func NewRedis(redisClient *redis.Client, now func() time.Time) *Redis {
	if redisClient == nil {
		panic(e.NewNilArgumentError("redisClient"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &Redis{redisClient: redisClient, now: now}
}

func (r *Redis) CheckLimit(ctx context.Context, key string, limit ratelimiter.Limit) error {
	var d time.Duration
	var k string

	switch limit.Interval {
	case ratelimiter.Hour:
		k = fmt.Sprintf("%s::h%d", key, r.now().Hour())
		d = time.Hour
	case ratelimiter.Minute:
		k = fmt.Sprintf("%s::m%d", key, r.now().Minute())
		d = time.Minute
	default:
		return fmt.Errorf("invalid rate limiting interval %v", limit.Interval)
	}

	cmds, err := r.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Incr(ctx, k)
		pipe.Expire(ctx, k, d)
		return nil
	})
	if err != nil {
		return err
	}
	intCmd := cmds[0].(*redis.IntCmd)
	if intCmd.Val() > int64(limit.Value) {
		return ratelimiter.NewRateLimiteExceededError(key)
	}
	return nil
}
