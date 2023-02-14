package user

import (
	"context"
	c "remindme/internal/core/domain/common"
	"time"
)

type CreateUserInput struct {
	Email           c.Optional[c.Email]
	PasswordHash    c.Optional[PasswordHash]
	Identity        c.Optional[Identity]
	CreatedAt       time.Time
	TimeZone        *time.Location
	ActivatedAt     c.Optional[time.Time]
	ActivationToken c.Optional[ActivationToken]
}

type UserRepository interface {
	Create(ctx context.Context, input CreateUserInput) (User, error)
	GetByID(ctx context.Context, id ID) (User, error)
	GetByEmail(ctx context.Context, email c.Email) (User, error)
	Activate(ctx context.Context, token ActivationToken, at time.Time) (User, error)
	SetPassword(ctx context.Context, id ID, password PasswordHash) error
}

type CreateSessionInput struct {
	UserID    ID
	Token     SessionToken
	CreatedAt time.Time
}

type SessionRepository interface {
	Create(ctx context.Context, input CreateSessionInput) error
	GetUserByToken(ctx context.Context, token SessionToken) (User, error)
	Delete(ctx context.Context, token SessionToken) (userID ID, err error)
}

type CreateLimitsInput struct {
	UserID ID
	Limits Limits
}

type LimitsRepository interface {
	Create(ctx context.Context, input CreateLimitsInput) (Limits, error)
	GetUserLimits(ctx context.Context, userID ID) (Limits, error)
	GetUserLimitsWithLock(ctx context.Context, userID ID) (Limits, error)
}
