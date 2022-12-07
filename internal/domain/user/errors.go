package user

import (
	"errors"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUserDoesNotExist   = errors.New("user does not exist")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
