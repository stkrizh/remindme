package changepassword

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
)

type Input struct {
	CurrentPassword user.RawPassword
	NewPassword     user.RawPassword
	User            user.User
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.User = u
	return i
}

type Result struct{}

type service struct {
	log            logging.Logger
	userRepository user.UserRepository
	passwordHasher user.PasswordHasher
}

func New(
	log logging.Logger,
	userRepository user.UserRepository,
	passwordHasher user.PasswordHasher,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if userRepository == nil {
		panic(e.NewNilArgumentError("userRepository"))
	}
	if passwordHasher == nil {
		panic(e.NewNilArgumentError("passwordHasher"))
	}
	return &service{
		log:            log,
		passwordHasher: passwordHasher,
		userRepository: userRepository,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	isCurrentPasswordValid := s.passwordHasher.ValidatePassword(
		input.CurrentPassword,
		input.User.PasswordHash.Value,
	)
	if !isCurrentPasswordValid {
		return result, user.ErrInvalidCredentials
	}

	newPasswordHash, err := s.passwordHasher.HashPassword(input.NewPassword)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("err", err))
		return result, err
	}
	if err := s.userRepository.SetPassword(ctx, input.User.ID, newPasswordHash); err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("err", err))
		return result, err
	}

	return Result{}, nil
}
