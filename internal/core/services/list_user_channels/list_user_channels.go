package listuserchannels

import (
	"context"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
)

type Input struct {
	UserID user.ID
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.UserID = u.ID
	return i
}

type Result struct {
	Channels []channel.Channel
}

type service struct {
	log               logging.Logger
	channelRepository channel.Repository
}

func New(
	log logging.Logger,
	channelRepository channel.Repository,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if channelRepository == nil {
		panic(e.NewNilArgumentError("channelRepository"))
	}
	return &service{
		log:               log,
		channelRepository: channelRepository,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	channels, err := s.channelRepository.Read(
		ctx,
		channel.ReadOptions{UserIDEquals: c.NewOptional(input.UserID, true)},
	)
	if err != nil {
		s.log.Error(
			ctx,
			"Could not read user channels.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}
	s.log.Info(
		ctx,
		"User channels successfully read.",
		logging.Entry("input", input),
		logging.Entry("channelCount", len(channels)),
	)
	return Result{Channels: channels}, nil
}
