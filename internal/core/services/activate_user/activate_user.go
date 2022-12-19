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
	log logging.Logger
	uow uow.UnitOfWork
	now func() time.Time
}

func New(
	log logging.Logger,
	uow uow.UnitOfWork,
	now func() time.Time,
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
		log: log,
		uow: uow,
		now: now,
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

	newChannel, err := uow.Channels().Create(
		ctx,
		channel.CreateInput{
			CreatedBy:  u.ID,
			Settings:   channel.NewEmailSettings(u.Email.Value),
			CreatedAt:  s.now(),
			VerifiedAt: c.NewOptional(s.now(), true),
		},
	)
	if err != nil {
		s.log.Error(
			ctx,
			"Could not create email channel for the activated user.",
			logging.Entry("user_id", u.ID),
			logging.Entry("email", u.Email),
			logging.Entry("err", err),
		)
		return result, err
	}
	s.log.Info(
		ctx,
		"Email channel successfully created for the activated user.",
		logging.Entry("user_id", u.ID),
		logging.Entry("email", u.Email),
		logging.Entry("channel_id", newChannel.ID),
	)

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
