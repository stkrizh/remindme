package common

import (
	"fmt"
	"strings"
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

type Email string

func NewEmail(rawEmail string) Email {
	return Email(strings.ToLower(rawEmail))
}
