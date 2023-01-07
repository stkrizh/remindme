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

type ReadOptions struct {
	CreatedByEquals c.Optional[user.ID]
	SentAfter       c.Optional[time.Time]
	StatusIn        c.Optional[[]Status]
	StatusNotEquals c.Optional[Status]
	OrderBy         OrderBy
	Limit           c.Optional[uint]
	Offset          uint
}

type UpdateInput struct {
	ID                  ID
	DoAtUpdate          bool
	At                  time.Time
	DoEveryUpdate       bool
	Every               c.Optional[Every]
	DoStatusUpdate      bool
	Status              Status
	DoScheduledAtUpdate bool
	ScheduledAt         c.Optional[time.Time]
	DoSentAtUpdate      bool
	SentAt              c.Optional[time.Time]
	DoCanceledAtUpdate  bool
	CanceledAt          c.Optional[time.Time]
}

type ReminderRepository interface {
	Create(ctx context.Context, input CreateInput) (Reminder, error)
	GetByID(ctx context.Context, id ID) (ReminderWithChannels, error)
	Read(ctx context.Context, options ReadOptions) ([]ReminderWithChannels, error)
	Count(ctx context.Context, options ReadOptions) (uint, error)
	Update(ctx context.Context, input UpdateInput) (Reminder, error)
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
