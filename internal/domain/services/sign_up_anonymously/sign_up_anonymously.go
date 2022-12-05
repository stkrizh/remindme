package signupanonymously

import (
	"context"
	"errors"
	"net/netip"
	"remindme/internal/domain/common"
	e "remindme/internal/domain/errors"
	"remindme/internal/domain/logging"
	uow "remindme/internal/domain/unit_of_work"
	"remindme/internal/domain/user"
	"time"
)

type Input struct {
	IP netip.Addr
}

type Result struct {
	User  user.User
	Token user.SessionToken
}

type service struct {
	log                   logging.Logger
	uow                   uow.UnitOfWork
	identityGenerator     user.IdentityGenerator
	sessionTokenGenerator user.SessionTokenGenerator
	now                   func() time.Time
}

func New(
	log logging.Logger,
	unitOfWork uow.UnitOfWork,
	identityGenerator user.IdentityGenerator,
	sessionTokenGenerator user.SessionTokenGenerator,
	now func() time.Time,
) *service {
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
	if now == nil {
		panic(e.NewNilArgumentError("now"))
	}
	return &service{
		log:                   log,
		uow:                   unitOfWork,
		identityGenerator:     identityGenerator,
		sessionTokenGenerator: sessionTokenGenerator,
		now:                   now,
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
	})
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
		s.log.Error(
			ctx,
			"Could not create anonymous user.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}

	sessionToken := s.sessionTokenGenerator.GenerateToken()
	err = uow.Sessions().Create(
		ctx,
		user.CreateSessionInput{
			UserID:    createdUser.ID,
			Token:     sessionToken,
			CreatedAt: now,
		},
	)
	if err != nil {
		s.log.Error(
			ctx,
			"Could not create session token for anonymous user.",
			logging.Entry("input", input),
			logging.Entry("err", err),
		)
		return result, err
	}

	err = uow.Commit(ctx)
	if errors.Is(err, context.Canceled) {
		return result, err
	}
	if err != nil {
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
		"Anonymous user has been created.",
		logging.Entry("id", createdUser.ID),
		logging.Entry("identity", createdUser.Identity),
		logging.Entry("ip", input.IP),
	)
	return Result{User: createdUser, Token: sessionToken}, nil
}
