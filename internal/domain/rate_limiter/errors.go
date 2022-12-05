package ratelimiter

import "fmt"

type RateLimitExceededError struct {
	key string
}

func NewRateLimiteExceededError(key string) *RateLimitExceededError {
	return &RateLimitExceededError{key: key}
}

func (e *RateLimitExceededError) Error() string {
	return fmt.Sprintf("rate limit exceeded for key '%s'", e.key)
}
