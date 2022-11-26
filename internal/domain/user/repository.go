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

type Repository interface {
	Create(ctx context.Context, input CreateUserInput) (*User, error)
	GetByID(ctx context.Context, id ID) (*User, error)
}
