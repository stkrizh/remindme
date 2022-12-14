package channel

import (
	"remindme/internal/core/domain/user"
	"time"
)

type ID int64

type Settings interface {
	Type() Type
}

type Channel struct {
	ID         ID
	Settings   Settings
	CreatedBy  user.ID
	CreatedAt  time.Time
	IsVerified bool
}

func (c Channel) Type() Type {
	return c.Settings.Type()
}
