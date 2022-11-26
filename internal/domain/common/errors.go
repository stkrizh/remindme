package common

type InvalidStateError struct {
	msg string
}

func NewInvalidStateError(msg string) *InvalidStateError {
	return &InvalidStateError{msg: msg}
}

func (e *InvalidStateError) Error() string {
	return e.msg
}
