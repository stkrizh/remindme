package reminder

import (
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"time"
)

type ID int64

type Reminder struct {
	Id         ID
	CreatedBy  user.ID
	At         time.Time
	Every      c.Optional[Every]
	CreatedAt  time.Time
	SentAt     c.Optional[time.Time]
	CanceledAt c.Optional[time.Time]
	Status     Status
}
