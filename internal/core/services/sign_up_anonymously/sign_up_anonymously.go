package signupanonymously

import (
	"context"
	"errors"
	"net/netip"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/logging"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
	"time"
)

type Input struct {
	IP       netip.Addr
	TimeZone *time.Location
}

type Result struct {
	User  user.User
	Token user.SessionToken
}

type service struct {
	log                           logging.Logger
	uow                           uow.UnitOfWork
	identityGenerator             user.IdentityGenerator
	sessionTokenGenerator         user.SessionTokenGenerator
	internalChannelTokenGenerator channel.InternalChannelTokenGenerator
	now                           func() time.Time
	defaultLimits                 user.Limits
}

func New(
	log logging.Logger,
	unitOfWork uow.UnitOfWork,
	identityGenerator user.IdentityGenerator,
	sessionTokenGenerator user.SessionTokenGenerator,
	internalChannelTokenGenerator channel.InternalChannelTokenGenerator,
	now func() time.Time,
	defaultLimits user.Limits,
) services.Service[Input, Result] {
	if unitOfWork == nil {
		panic(e.NewNilArgumentError("unitOfWork"))
	}
	if log == nil {
		panic(e.NewNilArgumentError("log"))
	}
	if identityGenerator == nil {
		panic(e.NewNilArgumentError("identityGenerator"))
	}
	if sessionTokenGenerator == nil {
		panic(e.NewNilArgumentError("sessionTokenGenerator"))
	}
	if internalChannelTokenGenerator == nil {
		panic(e.NewNilArgumentError("internalChannelTokenGenerator"))
	}
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:                           log,
		uow:                           unitOfWork,
		identityGenerator:             identityGenerator,
		sessionTokenGenerator:         sessionTokenGenerator,
		internalChannelTokenGenerator: internalChannelTokenGenerator,
		now:                           now,
		defaultLimits:                 defaultLimits,
	}
}

func (s *service) Run(ctx context.Context, input Input) (result Result, err error) {
	uow, err := s.uow.Begin(ctx)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
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

	now := s.now()
	createdUser, err := uow.Users().Create(ctx, user.CreateUserInput{
		Identity:    common.NewOptional(s.identityGenerator.GenerateIdentity(), true),
		CreatedAt:   now,
		ActivatedAt: common.NewOptional(now, true),
		TimeZone:    input.TimeZone,
	})
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("err", err))
		return result, err
	}

	_, err = uow.Limits().Create(
		ctx,
		user.CreateLimitsInput{
			UserID: createdUser.ID,
			Limits: s.defaultLimits,
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("userID", createdUser.ID), logging.Entry("err", err))
		return result, err
	}

	if err := s.createInternalChannel(ctx, uow, createdUser); err != nil {
		return result, err
	}

	sessionToken := s.sessionTokenGenerator.GenerateSessionToken()
	err = uow.Sessions().Create(
		ctx,
		user.CreateSessionInput{
			UserID:    createdUser.ID,
			Token:     sessionToken,
			CreatedAt: now,
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("err", err))
		return result, err
	}

	err = uow.Commit(ctx)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("input", input), logging.Entry("err", err))
		return result, err
	}

	s.log.Info(
		ctx,
		"Anonymous user has been created.",
		logging.Entry("id", createdUser.ID),
		logging.Entry("identity", createdUser.Identity),
		logging.Entry("ip", input.IP),
	)
	return Result{User: createdUser, Token: sessionToken}, nil
}

func (s *service) createInternalChannel(
	ctx context.Context,
	uow uow.Context,
	user user.User,
) error {
	now := s.now()
	token := s.internalChannelTokenGenerator.GenerateInternalChannelToken()
	newChannel, err := uow.Channels().Create(
		ctx,
		channel.CreateInput{
			CreatedBy:  user.ID,
			Type:       channel.Internal,
			Settings:   channel.NewInternalSettings(token),
			CreatedAt:  now,
			VerifiedAt: common.NewOptional(now, true),
		},
	)
	if err != nil {
		logging.Error(ctx, s.log, err, logging.Entry("userID", user.ID), logging.Entry("token", token))
		return err
	}
	s.log.Info(
		ctx,
		"Internal channel successfully created for the anonymous user.",
		logging.Entry("userID", user.ID),
		logging.Entry("token", token),
		logging.Entry("channelID", newChannel.ID),
	)
	return nil
}
