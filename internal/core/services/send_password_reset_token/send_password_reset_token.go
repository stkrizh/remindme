package sendpasswordresettoken

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
)

type Input struct {
	Email user.Email
}

type Result struct{}

type service struct {
	log              logging.Logger
	userRepository   user.UserRepository
	passwordResetter user.PasswordResetter
}

func New(
	log logging.Logger,
	userRepository user.UserRepository,
	passwordResetter user.PasswordResetter,
) *service {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if userRepository == nil {
		panic(e.NewNilArgumentError("userRepository"))
	}
	if passwordResetter == nil {
		panic(e.NewNilArgumentError("passwordResetter"))
	}
	return &service{
		log:              log,
		userRepository:   userRepository,
		passwordResetter: passwordResetter,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	return result, err
}
