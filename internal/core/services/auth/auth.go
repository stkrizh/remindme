package auth

import (
	"context"
	e "remindme/internal/core/domain/errors"
	"remindme/internal/core/domain/user"
	"remindme/internal/core/services"
)

type contextAuthToken string

const CONTEXT_AUTH_TOKEN_KEY = contextAuthToken("authToken")

type Input interface {
	WithAuthenticatedUser(u user.User) Input
}

type service[T Input, S any] struct {
	sessionRepository user.SessionRepository
	inner             services.Service[T, S]
}

func WithAuthentication[T Input, S any](
	sessionRepository user.SessionRepository,
	inner services.Service[T, S],
) services.Service[T, S] {
	if sessionRepository == nil {
		panic(e.NewNilArgumentError("sessionRepository"))
	}
	if inner == nil {
		panic(e.NewNilArgumentError("inner"))
	}
	return &service[T, S]{
		sessionRepository: sessionRepository,
		inner:             inner,
	}
}

func (s *service[T, S]) Run(ctx context.Context, input T) (result S, err error) {
	authToken, ok := ctx.Value(CONTEXT_AUTH_TOKEN_KEY).(user.SessionToken)
	if !ok {
		return result, user.ErrUserDoesNotExist
	}
	u, err := s.sessionRepository.GetUserByToken(ctx, authToken)
	if err != nil {
		return result, err
	}
	return s.inner.Run(ctx, input.WithAuthenticatedUser(u).(T))
}
