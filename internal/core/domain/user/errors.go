package user

import (
	"errors"
)

var (
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrUserDoesNotExist           = errors.New("user does not exist")
	ErrInvalidCredentials         = errors.New("invalid credentials")
	ErrUserIsNotActive            = errors.New("user is not active")
	ErrSessionDoesNotExist        = errors.New("session does not exist")
	ErrInvalidPasswordFResetToken = errors.New("invalid password reset token")
)

var (
	ErrLimitEmailChannelCountExceeded = errors.New("email channel count limit exceeded")
)
