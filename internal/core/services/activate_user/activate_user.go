package activateuser

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
	"time"
)

type Input struct {
	ActivationToken user.ActivationToken
}

type Result struct{}

type service struct {
	log           logging.Logger
	uow           uow.UnitOfWork
	now           func() time.Time
	defaultLimits user.Limits
}

func New(
	log logging.Logger,
	uow uow.UnitOfWork,
	now func() time.Time,
	defaultLimits user.Limits,
) services.Service[Input, Result] {
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if uow == nil {
		panic(e.NewNilArgumentError("uow"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:           log,
		uow:           uow,
		now:           now,
		defaultLimits: defaultLimits,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	uow, err := s.uow.Begin(ctx)
	if err != nil {
		s.log.Error(
			ctx,
			"Could not begin unit of work.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}
	defer uow.Rollback(ctx)

	u, err := uow.Users().Activate(ctx, input.ActivationToken, s.now())
	if errors.Is(err, user.ErrInvalidActivationToken) {
		return result, err
	}
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input))
		return result, err
	}
	s.log.Info(ctx, "User successfully activated.", logging.Entry("userId", u.ID))

	if err = s.createLimits(ctx, uow, u); err != nil {
		return result, err
	}
	if err = s.createInternalChannel(ctx, uow, u); err != nil {
		return result, err
	}
	if err = s.createEmailChannel(ctx, uow, u); err != nil {
		return result, err
	}

	if err = uow.Commit(ctx); err != nil {
		s.log.Error(
			ctx,
			"Could not commit unit of work.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}

	return Result{}, nil
}

func (s *service) createLimits(
	ctx context.Context,
	uow uow.Context,
	u user.User,
) error {
	_, err := uow.Limits().Create(
		ctx,
		user.CreateLimitsInput{
			UserID: u.ID,
			Limits: s.defaultLimits,
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("userID", u.ID))
		return err
	}
	s.log.Info(
		ctx,
		"Limits record successfully created for user.",
		logging.Entry("userID", u.ID),
	)
	return nil
}

func (s *service) createInternalChannel(
	ctx context.Context,
	uow uow.Context,
	user user.User,
) error {
	now := s.now()
	newChannel, err := uow.Channels().Create(
		ctx,
		channel.CreateInput{
			CreatedBy:  user.ID,
			Type:       channel.Internal,
			Settings:   channel.NewInternalSettings(),
			CreatedAt:  now,
			VerifiedAt: c.NewOptional(now, true),
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("userID", user.ID))
		return err
	}
	s.log.Info(
		ctx,
		"Internal channel successfully created for the activated user.",
		logging.Entry("userID", user.ID),
		logging.Entry("channelID", newChannel.ID),
	)
	return nil
}

func (s *service) createEmailChannel(
	ctx context.Context,
	uow uow.Context,
	user user.User,
) error {
	now := s.now()
	newChannel, err := uow.Channels().Create(
		ctx,
		channel.CreateInput{
			CreatedBy:  user.ID,
			Type:       channel.Email,
			Settings:   channel.NewEmailSettings(user.Email.Value),
			CreatedAt:  now,
			VerifiedAt: c.NewOptional(now, true),
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("userID", user.ID), logging.Entry("email", user.Email))
		return err
	}
	s.log.Info(
		ctx,
		"Email channel successfully created for the activated user.",
		logging.Entry("userID", user.ID),
		logging.Entry("email", user.Email),
		logging.Entry("channelID", newChannel.ID),
	)
	return nil
}
