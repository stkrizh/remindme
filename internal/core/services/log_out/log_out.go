package logout

import (
	"context"
	"errors"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
)

type Input struct {
	Token user.SessionToken
}

type Result struct{}

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
	userID, err := s.sessionRepository.Delete(ctx, input.Token)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if errors.Is(err, user.ErrSessionDoesNotExist) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not delete user session.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}

	s.log.Info(ctx, "User successfully logged out.", logging.Entry("userId", userID))
	return Result{}, nil
}
