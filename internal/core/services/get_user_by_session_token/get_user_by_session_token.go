package getuserbysessiontoken

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
)

type Input struct {
	Token user.SessionToken
}

type Result struct {
	User user.User
}

type service struct {
	log               logging.Logger
	sessionRepository user.SessionRepository
}

func New(
	log logging.Logger,
	sessionRepository user.SessionRepository,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if sessionRepository == nil {
		panic(e.NewNilArgumentError("sessionRepository"))
	}
	return &service{
		log:               log,
		sessionRepository: sessionRepository,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	u, err := s.sessionRepository.GetUserByToken(ctx, input.Token)
	return Result{User: u}, err
}
