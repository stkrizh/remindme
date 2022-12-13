package activateuser

import (
	"context"
	"errors"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"time"
)

type Input struct {
	ActivationToken user.ActivationToken
}

type Result struct{}

type service struct {
	log            logging.Logger
	userRepository user.UserRepository
	now            func() time.Time
}

func New(
	log logging.Logger,
	userRepository user.UserRepository,
	now func() time.Time,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if userRepository == nil {
		panic(e.NewNilArgumentError("userRepository"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:            log,
		userRepository: userRepository,
		now:            now,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	u, err := s.userRepository.Activate(ctx, input.ActivationToken, s.now())
	if errors.Is(err, user.ErrUserDoesNotExist) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not activate user.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}
	s.log.Info(ctx, "User successfully activated.", logging.Entry("userId", u.ID))
	return Result{}, nil
}
