package common

type Optional[T any] struct {
	Value     T
	IsPresent bool
}

func NewOptional[T any](value T, isPresent bool) Optional[T] {
	return Optional[T]{Value: value, IsPresent: isPresent}
}
