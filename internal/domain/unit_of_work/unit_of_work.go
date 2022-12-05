package uow

import (
	"context"
	"remindme/internal/domain/user"
)

type Context interface {
	Rollback(ctx context.Context) error
	Commit(ctx context.Context) error

	Users() user.UserRepository
	Sessions() user.SessionRepository
}

type UnitOfWork interface {
	Begin(ctx context.Context) (Context, error)
}
