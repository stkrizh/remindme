package errors

import "fmt"

type InvalidStateError struct {
	msg string
}

func NewInvalidStateError(msg string) *InvalidStateError {
	return &InvalidStateError{msg: msg}
}

func (e *InvalidStateError) Error() string {
	return e.msg
}

type NilArgumentError struct {
	argument string
}

func NewNilArgumentError(argument string) *NilArgumentError {
	return &NilArgumentError{argument: argument}
}

func (e *NilArgumentError) Error() string {
	return fmt.Sprintf("argument '%s' must not be nil", e.argument)
}
