package reminder

import (
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	e "remindme/internal/core/domain/errors"
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
	ChannelIDs map[channel.ID]struct{}
}

func (r *Reminder) Validate() error {
	if !r.Every.Value.IsValid() {
		return e.NewInvalidStateError("value of Every is not valid")
	}
	if r.SentAt.IsPresent && r.CanceledAt.IsPresent {
		return e.NewInvalidStateError("either SentAt or CanceledAt must not be set")
	}
	if r.Status == Sent && !r.SentAt.IsPresent {
		return e.NewInvalidStateError("SentAt must be set for sent reminders")
	}
	if r.Status == Canceled && !r.CanceledAt.IsPresent {
		return e.NewInvalidStateError("CanceledAt must be set for canceled reminders")
	}
	if len(r.ChannelIDs) == 0 {
		return e.NewInvalidStateError("reminder must have at least one channel")
	}
	return nil
}

type ReminderWithChannels struct {
	Reminder Reminder
	Channels []channel.Channel
}
