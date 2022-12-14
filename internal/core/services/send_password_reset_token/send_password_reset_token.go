package sendpasswordresettoken

import (
	"context"
	"errors"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
)

type Input struct {
	Email c.Email
}

func (i Input) GetRateLimitKey() string {
	return "send-password-reset-token::" + string(i.Email)
}

type Result struct {
	Token user.PasswordResetToken
}

type service struct {
	log              logging.Logger
	userRepository   user.UserRepository
	passwordResetter user.PasswordResetter
	sender           user.PasswordResetTokenSender
}

func New(
	log logging.Logger,
	userRepository user.UserRepository,
	passwordResetter user.PasswordResetter,
	sender user.PasswordResetTokenSender,
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
	if sender == nil {
		panic(e.NewNilArgumentError("sender"))
	}
	return &service{
		log:              log,
		userRepository:   userRepository,
		passwordResetter: passwordResetter,
		sender:           sender,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	u, err := s.userRepository.GetByEmail(ctx, input.Email)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if errors.Is(err, user.ErrUserDoesNotExist) {
		s.log.Info(ctx, "User not found for password reset.", logging.Entry("input", input))
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not get user for password reset.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}

	token := s.passwordResetter.GenerateToken(u)
	err = s.sender.SendToken(ctx, u, token)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not send password reset token.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}

	s.log.Info(
		ctx,
		"Password reset token has been successfully sent.",
		logging.Entry("input", input),
	)
	return Result{Token: token}, nil
}
