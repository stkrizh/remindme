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

type UpdateInput struct {
	ID                        ID
	DoVerificationTokenUpdate bool
	VerificationToken         c.Optional[VerificationToken]
	DoVerifiedAtUpdate        bool
	VerifiedAt                c.Optional[time.Time]
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Channel, error)
	GetByID(ctx context.Context, id ID) (Channel, error)
	Read(ctx context.Context, options ReadOptions) ([]Channel, error)
	Count(ctx context.Context, options ReadOptions) (uint, error)
	Update(ctx context.Context, input UpdateInput) (Channel, error)
}
