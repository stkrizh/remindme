package channel

import (
	"remindme/internal/core/domain/user"
	"time"
)

type ID int64

type Channel struct {
	ID         ID
	Settings   Settings
	CreatedBy  user.ID
	CreatedAt  time.Time
	IsVerified bool
}
