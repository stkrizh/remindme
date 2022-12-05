package signupwithemail

import (
	"context"
	"errors"
	c "remindme/internal/domain/common"
	e "remindme/internal/domain/errors"
	"remindme/internal/domain/logging"
	uow "remindme/internal/domain/unit_of_work"
	"remindme/internal/domain/user"
	"time"
)

type Input struct {
	Email    user.Email
	Password user.RawPassword
}

type Result struct {
	User user.User
}

type service struct {
	log                      logging.Logger
	unitOfWork               uow.UnitOfWork
	passwordHasher           user.PasswordHasher
	activationTokenGenerator user.ActivationTokenGenerator
	now                      func() time.Time
}

func New(
	log logging.Logger,
	unitOfWork uow.UnitOfWork,
	passwordHasher user.PasswordHasher,
	activationTokenGenerator user.ActivationTokenGenerator,
	now func() time.Time,
) *service {
	if unitOfWork == nil {
		panic(e.NewNilArgumentError("unitOfWork"))
	}
	if passwordHasher == nil {
		panic(e.NewNilArgumentError("passwordHasher"))
	}
	if activationTokenGenerator == nil {
		panic(e.NewNilArgumentError("activationTokenGenerator"))
	}
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		unitOfWork:               unitOfWork,
		passwordHasher:           passwordHasher,
		activationTokenGenerator: activationTokenGenerator,
		log:                      log,
		now:                      now,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	passwordHash, err := s.passwordHasher.HashPassword(input.Password)
	if err != nil {
		s.log.Error(ctx, "Could not hash password.", logging.Entry("err", err))
		return result, err
	}
	uow, err := s.unitOfWork.Begin(ctx)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not begin unit of work.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}
	defer uow.Rollback(ctx)

	createdUser, err := uow.Users().Create(ctx, user.CreateUserInput{
		Email:           c.NewOptional(input.Email, true),
		PasswordHash:    c.NewOptional(passwordHash, true),
		CreatedAt:       s.now(),
		ActivationToken: c.NewOptional(s.activationTokenGenerator.GenerateToken(), true),
	})
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		var emailAlreadyExistsErr *user.EmailAlreadyExistsError
		if errors.As(err, &emailAlreadyExistsErr) {
			s.log.Info(
				ctx,
				"User with the email already exists.",
				logging.Entry("email", input.Email),
			)
		} else {
			s.log.Error(
				ctx,
				"Could not create new user.",
				logging.Entry("input", input),
				logging.Entry("err", err),
			)
		}
		return result, err
	}

	err = uow.Commit(ctx)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not commit unit of work.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}

	s.log.Info(ctx, "New user has been created.", logging.Entry("user", createdUser))
	return Result{User: createdUser}, nil
}
