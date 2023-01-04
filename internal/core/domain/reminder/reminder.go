package reminder

import (
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"time"
)

const (
	MIN_DURATION_FROM_NOW   = 30 * time.Second
	DURATION_FOR_SCHEDULING = 24 * time.Hour
	MAX_CHANNEL_COUNT       = 5
)

type ID int64

type Reminder struct {
	ID          ID
	CreatedBy   user.ID
	At          time.Time
	Every       c.Optional[Every]
	CreatedAt   time.Time
	Status      Status
	ScheduledAt c.Optional[time.Time]
	SentAt      c.Optional[time.Time]
	CanceledAt  c.Optional[time.Time]
}

func (r *Reminder) Validate() error {
	if r.Every.IsPresent {
		if err := r.Every.Value.Validate(); err != nil {
			return err
		}
	}
	return nil
}

type ReminderWithChannels struct {
	Reminder
	Channels []channel.Channel
}

func (r *ReminderWithChannels) FromReminderAndChannels(reminder Reminder, channels []channel.Channel) {
	r.ID = reminder.ID
	r.CreatedBy = reminder.CreatedBy
	r.At = reminder.At
	r.Every = reminder.Every
	r.CreatedAt = reminder.CreatedAt
	r.SentAt = reminder.SentAt
	r.CanceledAt = reminder.CanceledAt
	r.Channels = channels
}
