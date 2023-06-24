package updateuser

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
	"time"
)

type Input struct {
	UserID           user.ID
	DoTimeZoneUpdate bool
	TimeZone         *time.Location
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.UserID = u.ID
	return i
}

type Result struct {
	User user.User
}

type service struct {
	log            logging.Logger
	userRepository user.UserRepository
}

func New(
	log logging.Logger,
	userRepository user.UserRepository,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if userRepository == nil {
		panic(e.NewNilArgumentError("userRepository"))
	}
	return &service{
		log:            log,
		userRepository: userRepository,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	updatedUser, err := s.userRepository.Update(
		ctx,
		user.UpdateUserInput{
			ID:               input.UserID,
			DoTimeZoneUpdate: input.DoTimeZoneUpdate,
			TimeZone:         input.TimeZone,
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}

	s.log.Info(
		ctx,
		"User successfully updated.",
		logging.Entry("input", input),
		logging.Entry("userID", updatedUser.ID),
	)
	result.User = updatedUser
	return result, nil
}
