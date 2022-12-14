package loginwithemail

import (
	"context"
	"errors"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"time"
)

type Input struct {
	Email    c.Email
	Password user.RawPassword
}

func (i Input) GetRateLimitKey() string {
	return "log-in-with-email::" + string(i.Email)
}

type Result struct {
	Token user.SessionToken
}

type service struct {
	log                   logging.Logger
	userRepository        user.UserRepository
	sessionRepository     user.SessionRepository
	passwordHasher        user.PasswordHasher
	sessionTokenGenerator user.SessionTokenGenerator
	now                   func() time.Time
}

func New(
	log logging.Logger,
	userRepository user.UserRepository,
	sessionRepository user.SessionRepository,
	passwordHasher user.PasswordHasher,
	sessionTokenGenerator user.SessionTokenGenerator,
	now func() time.Time,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if userRepository == nil {
		panic(e.NewNilArgumentError("userRepository"))
	}
	if sessionRepository == nil {
		panic(e.NewNilArgumentError("sessionRepository"))
	}
	if passwordHasher == nil {
		panic(e.NewNilArgumentError("passwordHasher"))
	}
	if sessionTokenGenerator == nil {
		panic(e.NewNilArgumentError("sessionTokenGenerator"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:                   log,
		userRepository:        userRepository,
		sessionRepository:     sessionRepository,
		passwordHasher:        passwordHasher,
		sessionTokenGenerator: sessionTokenGenerator,
		now:                   now,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	u, err := s.userRepository.GetByEmail(ctx, input.Email)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if errors.Is(err, user.ErrUserDoesNotExist) {
		// Minimize risk for timing attacks
		s.passwordHasher.HashPassword(input.Password)
		return result, user.ErrInvalidCredentials
	}
	if !s.passwordHasher.ValidatePassword(input.Password, u.PasswordHash.Value) {
		return result, user.ErrInvalidCredentials
	}
	if !u.IsActive() {
		return result, user.ErrUserIsNotActive
	}

	sessionToken := s.sessionTokenGenerator.GenerateToken()
	err = s.sessionRepository.Create(ctx, user.CreateSessionInput{
		UserID:    u.ID,
		Token:     sessionToken,
		CreatedAt: s.now(),
	})
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not create session token for user.",
			logging.Entry("userId", u.ID),
			logging.Entry("err", err),
		)
		return result, err
	}

	s.log.Info(
		ctx,
		"User successfully authenticated, session token created.",
		logging.Entry("userId", u.ID),
	)
	return Result{Token: sessionToken}, nil
}
