package ratelimiter

import (
	"context"
	"errors"
	"fmt"
	e "remindme/internal/domain/errors"
	"remindme/internal/domain/logging"
	ratelimiter "remindme/internal/domain/rate_limiter"
	"time"

	"github.com/go-redis/redis/v9"
)

type Redis struct {
	redisClient *redis.Client
	log         logging.Logger
	now         func() time.Time
}

func NewRedis(redisClient *redis.Client, log logging.Logger, now func() time.Time) *Redis {
	if redisClient == nil {
		panic(e.NewNilArgumentError("redisClient"))
	}
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &Redis{redisClient: redisClient, log: log, now: now}
}

func (r *Redis) CheckLimit(ctx context.Context, key string, limit ratelimiter.Limit) ratelimiter.Result {
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
		panic("invalid rate limiting interval")
	}

	cmds, err := r.redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Incr(ctx, k)
		pipe.Expire(ctx, k, d)
		return nil
	})
	if errors.Is(err, context.Canceled) {
		return ratelimiter.NotAllowed()
	}
	if err != nil {
		r.log.Error(ctx, "Could not check rate limit due to Redis client error.", logging.Entry("err", err))
		return ratelimiter.Allowed()
	}
	intCmd := cmds[0].(*redis.IntCmd)
	if intCmd.Val() > int64(limit.Value) {
		return ratelimiter.NotAllowed()
	}
	return ratelimiter.Allowed()
}
