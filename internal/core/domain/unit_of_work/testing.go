package uow

import (
	"context"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/reminder"
	"remindme/internal/core/domain/user"
)

type FakeUnitOfWorkContext struct {
	UserRepository            *user.FakeUserRepository
	SessionRepository         *user.FakeSessionRepository
	LimitsRepository          *user.FakeLimitsRepository
	ChannelRepository         *channel.FakeRepository
	ReminderRepository        *reminder.FakeReminderRepository
	ReminderChannelRepository *reminder.FakeReminderChannelRepository
	WasRollbackCalled         bool
	WasCommitCalled           bool
}

func NewFakeUnitOfWorkContext(
	userRepository *user.FakeUserRepository,
	sessionRepository *user.FakeSessionRepository,
	limitsRepository *user.FakeLimitsRepository,
	channelRepository *channel.FakeRepository,
	reminderRepository *reminder.FakeReminderRepository,
	reminderChannelRepository *reminder.FakeReminderChannelRepository,
) *FakeUnitOfWorkContext {
	return &FakeUnitOfWorkContext{
		UserRepository:            userRepository,
		SessionRepository:         sessionRepository,
		LimitsRepository:          limitsRepository,
		ChannelRepository:         channelRepository,
		ReminderRepository:        reminderRepository,
		ReminderChannelRepository: reminderChannelRepository,
	}
}

func (c *FakeUnitOfWorkContext) Rollback(ctx context.Context) error {
	c.WasRollbackCalled = true
	return nil
}

func (c *FakeUnitOfWorkContext) Commit(ctx context.Context) error {
	c.WasCommitCalled = true
	return nil
}

func (c *FakeUnitOfWorkContext) Users() user.UserRepository {
	return c.UserRepository
}

func (c *FakeUnitOfWorkContext) Sessions() user.SessionRepository {
	return c.SessionRepository
}

func (c *FakeUnitOfWorkContext) Limits() user.LimitsRepository {
	return c.LimitsRepository
}

func (c *FakeUnitOfWorkContext) Channels() channel.Repository {
	return c.ChannelRepository
}

func (c *FakeUnitOfWorkContext) Reminders() reminder.ReminderRepository {
	return c.ReminderRepository
}

func (c *FakeUnitOfWorkContext) ReminderChannels() reminder.ReminderChannelRepository {
	return c.ReminderChannelRepository
}

type FakeUnitOfWork struct {
	Context *FakeUnitOfWorkContext
}

func NewFakeUnitOfWork() *FakeUnitOfWork {
	userRepository := user.NewFakeUserRepository()
	return &FakeUnitOfWork{
		Context: NewFakeUnitOfWorkContext(
			userRepository,
			user.NewFakeSessionRepository(userRepository),
			user.NewFakeLimitsRepository(),
			channel.NewFakeRepository(),
			reminder.NewFakeReminderRepository(),
			reminder.NewFakeReminderChannelRepository(),
		),
	}
}

func (u *FakeUnitOfWork) Begin(ctx context.Context) (Context, error) {
	return u.Context, nil
}

func (u *FakeUnitOfWork) Users() *user.FakeUserRepository {
	return u.Context.UserRepository
}

func (u *FakeUnitOfWork) Sessions() *user.FakeSessionRepository {
	return u.Context.SessionRepository
}

func (u *FakeUnitOfWork) Limits() *user.FakeLimitsRepository {
	return u.Context.LimitsRepository
}

func (u *FakeUnitOfWork) Channels() *channel.FakeRepository {
	return u.Context.ChannelRepository
}

func (u *FakeUnitOfWork) Reminders() *reminder.FakeReminderRepository {
	return u.Context.ReminderRepository
}

func (u *FakeUnitOfWork) ReminderChannels() *reminder.FakeReminderChannelRepository {
	return u.Context.ReminderChannelRepository
}
