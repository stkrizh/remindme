package signinwithemail

import (
	"context"
	"errors"
	e "remindme/internal/domain/errors"
	"remindme/internal/domain/logging"
	"remindme/internal/domain/user"
	"time"
)

type Input struct {
	Email    user.Email
	Password user.RawPassword
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
) *service {
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
	if errors.Is(err, user.ErrUserDoesNotExist) {
		// Minimize risk for timing attacks
		s.passwordHasher.HashPassword(input.Password)
		return result, user.ErrInvalidCredentials
	}
	if !s.passwordHasher.ValidatePassword(input.Password, u.PasswordHash.Value) {
		return result, user.ErrInvalidCredentials
	}
	sessionToken := s.sessionTokenGenerator.GenerateToken()
	err = s.sessionRepository.Create(ctx, user.CreateSessionInput{
		UserID:    u.ID,
		Token:     sessionToken,
		CreatedAt: s.now(),
	})
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
