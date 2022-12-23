package createemailchannel

import (
	"context"
	"errors"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"remindme/internal/core/services/auth"
	"time"
)

type Input struct {
	Email  c.Email
	UserID user.ID
}

func (i Input) WithAuthenticatedUser(u user.User) auth.Input {
	i.UserID = u.ID
	return i
}

type Result struct {
	Channel           channel.Channel
	VerificationToken channel.VerificationToken
}

type service struct {
	log            logging.Logger
	unitOfWork     uow.UnitOfWork
	tokenGenerator channel.VerificationTokenGenerator
	now            func() time.Time
}

func New(
	log logging.Logger,
	unitOfWork uow.UnitOfWork,
	tokenGenerator channel.VerificationTokenGenerator,
	now func() time.Time,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if unitOfWork == nil {
		panic(e.NewNilArgumentError("unitOfWork"))
	}
	if tokenGenerator == nil {
		panic(e.NewNilArgumentError("tokenGenerator"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:            log,
		unitOfWork:     unitOfWork,
		tokenGenerator: tokenGenerator,
		now:            now,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	channelSettings := channel.NewEmailSettings(input.Email)
	token := s.tokenGenerator.GenerateVerificationToken()

	uow, err := s.unitOfWork.Begin(ctx)
	if err != nil {
		s.log.Error(
			ctx,
			"Could not begin unit of work.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
	}
	defer uow.Rollback(ctx)

	userLimits, err := uow.Limits().GetUserLimitsWithLock(ctx, input.UserID)
	if err != nil {
		s.log.Error(
			ctx,
			"Could not get user limits.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}
	actualEmailChannelCount, err := uow.Channels().Count(
		ctx,
		channel.ReadOptions{
			UserIDEquals: c.NewOptional(input.UserID, true),
			TypeEquals:   c.NewOptional(channel.Email, true),
		},
	)
	if err != nil {
		s.log.Error(
			ctx,
			"Could not get email channel count.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}
	if userLimits.EmailChannelCount.IsPresent && actualEmailChannelCount >= uint(userLimits.EmailChannelCount.Value) {
		return result, user.ErrLimitEmailChannelCountExceeded
	}

	newChannel, err := uow.Channels().Create(
		ctx,
		channel.CreateInput{
			CreatedBy:         input.UserID,
			Type:              channel.Email,
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
			logging.Entry("userID", input.UserID),
			logging.Entry("err", err),
		)
		return result, err
	}

	if err := uow.Commit(ctx); err != nil {
		s.log.Error(
			ctx,
			"Could not commit unit of work.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}
	s.log.Info(
		ctx,
		"New email channel has been created.",
		logging.Entry("email", input.Email),
		logging.Entry("userID", input.UserID),
		logging.Entry("channelID", newChannel.ID),
	)
	return Result{Channel: newChannel, VerificationToken: token}, nil
}
