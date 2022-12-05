package common

import (
	"fmt"
)

type Optional[T any] struct {
	Value     T
	IsPresent bool
}

func (p *Optional[T]) String() string {
	if !p.IsPresent {
		return "[-]"
	}
	return fmt.Sprintf("[%v]", p.Value)
}

func NewOptional[T any](value T, isPresent bool) Optional[T] {
	return Optional[T]{Value: value, IsPresent: isPresent}
}
