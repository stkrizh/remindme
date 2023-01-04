package uow

import (
	"context"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
)

type Context interface {
	Rollback(ctx context.Context) error
	Commit(ctx context.Context) error

	Users() user.UserRepository
	Sessions() user.SessionRepository
	Limits() user.LimitsRepository
	Channels() channel.Repository
	Reminders() reminder.ReminderRepository
	ReminderChannels() reminder.ReminderChannelRepository
}

type UnitOfWork interface {
	Begin(ctx context.Context) (Context, error)
}
