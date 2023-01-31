package reminder

import (
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"time"
)

const (
	MIN_DURATION_FROM_NOW   = 30 * time.Second
	MAX_DURATION_FROM_NOW   = 2 * 366 * 24 * time.Hour
	DURATION_FOR_SCHEDULING = 24 * time.Hour
	MAX_CHANNEL_COUNT       = 5
	MAX_BODY_LEN            = 280
	MAX_SENDING_DELAY       = 10 * time.Minute
)

type ID int64

type Reminder struct {
	ID          ID
	CreatedBy   user.ID
	At          time.Time
	Body        string
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

func (r *Reminder) IsActive() bool {
	return r.Status == StatusCreated || r.Status == StatusScheduled
}

type ReminderWithChannels struct {
	Reminder
	ChannelIDs []channel.ID
}

func (r *ReminderWithChannels) FromReminderAndChannels(reminder Reminder, channelIDs []channel.ID) {
	r.ID = reminder.ID
	r.CreatedBy = reminder.CreatedBy
	r.Status = reminder.Status
	r.At = reminder.At
	r.Body = reminder.Body
	r.Every = reminder.Every
	r.CreatedAt = reminder.CreatedAt
	r.ScheduledAt = reminder.ScheduledAt
	r.SentAt = reminder.SentAt
	r.CanceledAt = reminder.CanceledAt
	r.ChannelIDs = channelIDs
}
