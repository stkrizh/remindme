package reminder

import (
	"remindme/internal/domain/user"
	"time"
)

type ID int64

type Reminder struct {
	Id        ID
	CreatedAt time.Time
	CreatedBy user.ID
}
