package channel

import (
	"context"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"time"
)

type CreateInput struct {
	CreatedBy         user.ID
	Type              Type
	Settings          Settings
	CreatedAt         time.Time
	VerificationToken c.Optional[VerificationToken]
	VerifiedAt        c.Optional[time.Time]
}

type ReadOptions struct {
	UserIDEquals c.Optional[user.ID]
	TypeEquals   c.Optional[Type]
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Channel, error)
	Read(ctx context.Context, options ReadOptions) ([]Channel, error)
	Count(ctx context.Context, options ReadOptions) (uint, error)
}
