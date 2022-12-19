package createemailchannel

import (
	"context"
	"errors"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"time"
)

type Input struct {
	Email c.Email
	User  user.User
}

type Result struct {
	Channel           channel.Channel
	VerificationToken channel.VerificationToken
}

type service struct {
	log               logging.Logger
	channelRepository channel.Repository
	tokenGenerator    channel.VerificationTokenGenerator
	now               func() time.Time
}

func New(
	log logging.Logger,
	channelRepository channel.Repository,
	tokenGenerator channel.VerificationTokenGenerator,
	now func() time.Time,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if channelRepository == nil {
		panic(e.NewNilArgumentError("channelRepository"))
	}
	if tokenGenerator == nil {
		panic(e.NewNilArgumentError("tokenGenerator"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:               log,
		channelRepository: channelRepository,
		tokenGenerator:    tokenGenerator,
		now:               now,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	channelSettings := channel.NewEmailSettings(input.Email)
	token := s.tokenGenerator.GenerateToken()
	newChannel, err := s.channelRepository.Create(
		ctx,
		channel.CreateInput{
			CreatedBy:         input.User.ID,
			Settings:          channelSettings,
			CreatedAt:         s.now(),
			VerificationToken: c.NewOptional(token, true),
		},
	)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not create email channel.",
			logging.Entry("email", input.Email),
			logging.Entry("userID", input.User.ID),
			logging.Entry("err", err),
		)
		return result, err
	}

	s.log.Info(
		ctx,
		"New email channel has been created.",
		logging.Entry("email", input.Email),
		logging.Entry("userID", input.User.ID),
		logging.Entry("channelID", newChannel.ID),
	)
	return Result{Channel: newChannel, VerificationToken: token}, nil
}
