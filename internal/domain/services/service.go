package services

import "context"

type Service[T any, S any] interface {
	Run(ctx context.Context, input T) (S, error)
}
