package user

import (
	"context"
	c "remindme/internal/domain/common"
	"time"
)

type CreateUserInput struct {
	Email           c.Optional[Email]
	PasswordHash    c.Optional[PasswordHash]
	Identity        c.Optional[Identity]
	CreatedAt       time.Time
	ActivatedAt     c.Optional[time.Time]
	ActivationToken c.Optional[ActivationToken]
}

type UserRepository interface {
	Create(ctx context.Context, input CreateUserInput) (User, error)
	GetByID(ctx context.Context, id ID) (User, error)
}

type CreateSessionInput struct {
	UserID    ID
	Token     SessionToken
	CreatedAt time.Time
}

type SessionRepository interface {
	Create(ctx context.Context, input CreateSessionInput) error
	GetUserByToken(ctx context.Context, token SessionToken) (User, error)
}
