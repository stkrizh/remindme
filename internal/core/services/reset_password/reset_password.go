package reset_password

import (
	"context"
	"errors"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
)

type Input struct {
	Token       user.PasswordResetToken
	NewPassword user.RawPassword
}

type Result struct{}

type service struct {
	log              logging.Logger
	userRepository   user.UserRepository
	passwordResetter user.PasswordResetter
	passwordHasher   user.PasswordHasher
}

func New(
	log logging.Logger,
	userRepository user.UserRepository,
	passwordResetter user.PasswordResetter,
	passwordHasher user.PasswordHasher,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if userRepository == nil {
		panic(e.NewNilArgumentError("userRepository"))
	}
	if passwordResetter == nil {
		panic(e.NewNilArgumentError("passwordResetter"))
	}
	if passwordHasher == nil {
		panic(e.NewNilArgumentError("passwordHasher"))
	}
	return &service{
		log:              log,
		userRepository:   userRepository,
		passwordResetter: passwordResetter,
		passwordHasher:   passwordHasher,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	userID, ok := s.passwordResetter.GetUserID(input.Token)
	if !ok {
		return result, user.ErrInvalidPasswordFResetToken
	}
	u, err := s.userRepository.GetByID(ctx, userID)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if errors.Is(err, user.ErrUserDoesNotExist) {
		s.log.Info(ctx, "User not found for password reset.", logging.Entry("userID", userID))
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not get user for password reset.",
			logging.Entry("userID", userID),
			logging.Entry("err", err),
		)
		return result, err
	}

	isValid := s.passwordResetter.ValidateToken(u, input.Token)
	if !isValid {
		return result, user.ErrInvalidPasswordFResetToken
	}

	newPasswordHash, err := s.passwordHasher.HashPassword(input.NewPassword)
	if err != nil {
		return result, err
	}
	err = s.userRepository.SetPassword(ctx, u.ID, newPasswordHash)
	if errors.Is(err, user.ErrUserDoesNotExist) {
		s.log.Info(ctx, "Could not update user password, user does not exist.", logging.Entry("userID", userID))
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not update user password.",
			logging.Entry("userID", userID),
			logging.Entry("err", err),
		)
		return result, err
	}

	s.log.Info(
		ctx,
		"New password has been successfully set.",
		logging.Entry("userID", userID),
	)
	return result, nil
}
