package channel

import (
	"context"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"time"
)

type CreateInput struct {
	CreatedBy         user.ID
	Settings          Settings
	CreatedAt         time.Time
	VerificationToken c.Optional[VerificationToken]
	VerifiedAt        c.Optional[time.Time]
}

type Repository interface {
	Create(ctx context.Context, input CreateInput) (Channel, error)
}
