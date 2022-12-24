package verifychannel

import (
	"context"
	"errors"
	"fmt"
	"remindme/internal/core/domain/channel"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
	"time"
)

type Input struct {
	ChannelID         channel.ID
	VerificationToken channel.VerificationToken
	UserID            user.ID
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.UserID = u.ID
	return i
}

func (i Input) GetRateLimitKey() string {
	return fmt.Sprintf("verify-channel::%d", i.UserID)
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
	activatedChannel, err := s.channelRepository.Verify(
		ctx,
		channel.VerifyInput{
			ID:                input.ChannelID,
			CreatedBy:         input.UserID,
			VerificationToken: input.VerificationToken,
			At:                s.now(),
		},
	)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			// Do nothing
		case errors.Is(err, channel.ErrChannelDoesNotExist):
			s.log.Info(ctx, "Invalid channel verification data.", logging.Entry("input", input))
		default:
			s.log.Error(
				ctx,
				"Could not verify channel due to unexpected error.",
				logging.Entry("input", input),
				logging.Entry("err", err),
			)
		}
		return result, err
	}

	s.log.Info(
		ctx,
		"Channel successfully verified.",
		logging.Entry("userID", input.UserID),
		logging.Entry("channelID", activatedChannel.ID),
		logging.Entry("channelType", activatedChannel.Type),
	)
	return Result{Channel: activatedChannel}, nil
}
