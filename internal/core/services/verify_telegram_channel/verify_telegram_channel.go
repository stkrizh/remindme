package verifytelegramchannel

import (
	"context"
	"errors"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/services"
	"time"
)

type Input struct {
	ChannelID         channel.ID
	VerificationToken channel.VerificationToken
	TelegramChatID    channel.TelegramChatID
}

type Result struct {
	Channel channel.Channel
}

type service struct {
	log               logging.Logger
	channelRepository channel.Repository
	now               func() time.Time
}

func New(
	log logging.Logger,
	channelRepository channel.Repository,
	now func() time.Time,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if channelRepository == nil {
		panic(e.NewNilArgumentError("channelRepository"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:               log,
		channelRepository: channelRepository,
		now:               now,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	existingChannel, err := s.channelRepository.GetByID(ctx, input.ChannelID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			// Do nothing
		case errors.Is(err, channel.ErrChannelDoesNotExist):
			s.log.Info(ctx, "Channel does not exist.", logging.Entry("input", input))
		default:
			s.log.Error(
				ctx,
				"Could not get channel for verification due to unexpected error.",
				logging.Entry("input", input),
				logging.Entry("err", err),
			)
		}
		return result, err
	}

	if !(existingChannel.Type == channel.Telegram &&
		existingChannel.VerificationToken.IsPresent &&
		existingChannel.VerificationToken.Value == input.VerificationToken) {
		s.log.Info(
			ctx,
			"Invalid telegram channel verification data.",
			logging.Entry("input", input),
			logging.Entry("channel", existingChannel),
		)
		return result, channel.ErrInvalidVerificationData
	}
	existingSettings := existingChannel.Settings.(*channel.TelegramSettings)

	verifiedChannel, err := s.channelRepository.Update(
		ctx,
		channel.UpdateInput{
			ID:                        input.ChannelID,
			DoVerificationTokenUpdate: true,
			VerificationToken:         c.NewOptional(channel.VerificationToken(""), false),
			DoVerifiedAtUpdate:        true,
			VerifiedAt:                c.NewOptional(s.now(), true),
			DoSettingsUpdate:          true,
			Settings:                  channel.NewTelegramSettings(existingSettings.Bot, input.TelegramChatID),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			// Do nothing
		case errors.Is(err, channel.ErrChannelDoesNotExist):
			s.log.Info(ctx, "Could not verify telegram channel, channel does not exist.", logging.Entry("input", input))
		default:
			s.log.Error(
				ctx,
				"Could not verify telegram channel due to unexpected error.",
				logging.Entry("input", input),
				logging.Entry("err", err),
			)
		}
		return result, err
	}

	s.log.Info(
		ctx,
		"Telegram channel successfully verified.",
		logging.Entry("channelID", verifiedChannel.ID),
		logging.Entry("channelType", verifiedChannel.Type),
		logging.Entry("telegramChatID", input.TelegramChatID),
	)
	return Result{Channel: verifiedChannel}, nil
}
