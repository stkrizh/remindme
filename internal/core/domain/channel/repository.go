package channel

import (
	"context"
	"remindme/internal/core/domain/user"
	"time"
)

type CreateInput struct {
	CreatedBy  user.ID
	Settings   Settings
	CreatedAt  time.Time
	IsVerified bool
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Channel, error)
}
