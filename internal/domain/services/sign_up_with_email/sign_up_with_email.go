package signupwithemail

import (
	"context"
	"errors"
	c "remindme/internal/domain/common"
	"remindme/internal/domain/logging"
	"remindme/internal/domain/services"
	uow "remindme/internal/domain/unit_of_work"
	"remindme/internal/domain/user"
	"time"
)

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
) services.Service[services.SignUpWithEmailInput, services.SignUpWithEmailResult] {
	if unitOfWork == nil {
		panic("Argument unitOfWork must not be nil.")
	}
	if passwordHasher == nil {
		panic("Argument passwordHasher must not be nil.")
	}
	if activationTokenGenerator == nil {
		panic("Argument activationTokenGenerator must not be nil.")
	}
	if log == nil {
		panic("Argument logger must not be nil.")
	}
	return &service{
		unitOfWork:               unitOfWork,
		passwordHasher:           passwordHasher,
		activationTokenGenerator: activationTokenGenerator,
		log:                      log,
		now:                      now,
	}
}

func (s *service) Run(
	ctx context.Context,
	input services.SignUpWithEmailInput,
) (result services.SignUpWithEmailResult, err error) {
	passwordHash, err := s.passwordHasher.HashPassword(input.Password)
	if err != nil {
		s.log.Error(ctx, "Could not hash password.", logging.Entry("err", err))
		return result, err
	}
	uow, err := s.unitOfWork.Begin(ctx)
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
	var emailAlreadyExistsErr *user.EmailAlreadyExistsError
	if errors.As(err, &emailAlreadyExistsErr) {
		s.log.Info(
			ctx,
			"User with the email already exists.",
			logging.Entry("email", input.Email),
		)
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not create new user.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}

	if err = uow.Commit(ctx); err != nil {
		s.log.Error(
			ctx,
			"Could not commit unit of work.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}
	s.log.Info(ctx, "New user has been created.", logging.Entry("user", createdUser))
	return services.SignUpWithEmailResult{User: createdUser}, nil
}
