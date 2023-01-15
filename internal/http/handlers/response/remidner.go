package response

import (
	"remindme/internal/core/domain/reminder"
	"time"
)

type ReminderWithChannels struct {
	ID          int64      `json:"id"`
	CreatedBy   int64      `json:"created_by"`
	At          time.Time  `json:"at"`
	Every       *string    `json:"every"`
	Body        string     `json:"body"`
	CreatedAt   time.Time  `json:"created_at"`
	Status      string     `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at"`
	SentAt      *time.Time `json:"sent_at"`
	CanceledAt  *time.Time `json:"canceled_at"`
	ChannelIDs  []int64    `json:"channel_ids"`
}

func (r *ReminderWithChannels) FromDomainType(dr reminder.ReminderWithChannels) {
	r.ID = int64(dr.ID)
	r.CreatedBy = int64(dr.CreatedBy)
	r.At = dr.At
	if dr.Every.IsPresent {
		every := dr.Every.Value.String()
		r.Every = &every
	}
	r.Body = dr.Body
	r.CreatedAt = dr.CreatedAt
	r.Status = dr.Status.String()
	if dr.ScheduledAt.IsPresent {
		r.ScheduledAt = &dr.ScheduledAt.Value
	}
	if dr.SentAt.IsPresent {
		r.SentAt = &dr.SentAt.Value
	}
	if dr.CanceledAt.IsPresent {
		r.CanceledAt = &dr.CanceledAt.Value
	}
	r.ChannelIDs = make([]int64, 0, len(dr.ChannelIDs))
	for _, channelID := range dr.ChannelIDs {
		r.ChannelIDs = append(r.ChannelIDs, int64(channelID))
	}
}
