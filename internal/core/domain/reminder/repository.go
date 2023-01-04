package reminder

import (
	"context"
	"remindme/internal/core/domain/channel"
	c "remindme/internal/core/domain/common"
	"remindme/internal/core/domain/user"
	"time"
)

type CreateInput struct {
	CreatedBy   user.ID
	CreatedAt   time.Time
	At          time.Time
	Every       c.Optional[Every]
	ScheduledAt c.Optional[time.Time]
	SentAt      c.Optional[time.Time]
	CanceledAt  c.Optional[time.Time]
	Status      Status
}

type OrderBy struct {
	v int
}

var (
	OrderByIDAsc  OrderBy = OrderBy{}
	OrderByIDDesc OrderBy = OrderBy{v: 1}
	OrderByAtAsc  OrderBy = OrderBy{v: 2}
	OrderByAtDesc OrderBy = OrderBy{v: 3}
)

type ReadOptions struct {
	CreatedByEquals c.Optional[user.ID]
	SentAfter       c.Optional[time.Time]
	StatusIn        c.Optional[[]Status]
	StatusNotEquals c.Optional[Status]
	OrderBy         OrderBy
}

type ReminderRepository interface {
	Create(ctx context.Context, input CreateInput) (Reminder, error)
	Read(ctx context.Context, options ReadOptions) ([]Reminder, error)
	Count(ctx context.Context, options ReadOptions) (uint, error)
}

type ChannelIDs map[channel.ID]struct{}

func NewChannelIDs(ids ...channel.ID) ChannelIDs {
	channelIDs := make(map[channel.ID]struct{}, len(ids))
	for _, id := range ids {
		channelIDs[id] = struct{}{}
	}
	return channelIDs
}

type CreateChannelsInput struct {
	ReminderID ID
	ChannelIDs ChannelIDs
}

func NewCreateChannelsInput(reminderID ID, channelIDs ...channel.ID) CreateChannelsInput {
	return CreateChannelsInput{
		ReminderID: reminderID,
		ChannelIDs: NewChannelIDs(channelIDs...),
	}
}

type ReminderChannelRepository interface {
	Create(ctx context.Context, input CreateChannelsInput) (ChannelIDs, error)
}
