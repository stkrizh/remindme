package getlimitforchannels

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
	Email    c.Optional[user.Limit]
	Telegram c.Optional[user.Limit]
}

type service struct {
	log      logging.Logger
	limits   user.LimitsRepository
	channels channel.Repository
}

func New(
	log logging.Logger,
	limits user.LimitsRepository,
	channels channel.Repository,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if limits == nil {
		panic(e.NewNilArgumentError("limits"))
	}
	if channels == nil {
		panic(e.NewNilArgumentError("channels"))
	}
	return &service{
		log:      log,
		limits:   limits,
		channels: channels,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	defer func() {
		if err != nil {
			logging.Error(ctx, s.log, err, logging.Entry("input", input))
		}
	}()

	limits, err := s.limits.GetUserLimits(ctx, input.UserID)
	if err != nil {
		return result, err
	}

	if !(limits.EmailChannelCount.IsPresent || limits.TelegramChannelCount.IsPresent) {
		return result, nil
	}

	channels, err := s.channels.Read(ctx, channel.ReadOptions{UserIDEquals: c.NewOptional(input.UserID, true)})
	if err != nil {
		return result, err
	}
	s.countChannels(limits, channels, &result)

	s.log.Info(
		ctx,
		"Got user limit for channels.",
		logging.Entry("userID", input.UserID),
		logging.Entry("result", result),
	)
	return result, err
}

func (s *service) countChannels(limits user.Limits, channels []channel.Channel, result *Result) {
	if limits.EmailChannelCount.IsPresent {
		result.Email.IsPresent = true
		result.Email.Value.Value = limits.EmailChannelCount.Value
	}
	if limits.TelegramChannelCount.IsPresent {
		result.Telegram.IsPresent = true
		result.Telegram.Value.Value = limits.TelegramChannelCount.Value
	}
	for _, ch := range channels {
		switch {
		case ch.Type == channel.Email && limits.EmailChannelCount.IsPresent:
			result.Email.Value.Actual++
		case ch.Type == channel.Telegram && limits.TelegramChannelCount.IsPresent:
			result.Telegram.Value.Actual++
		}
	}
}
