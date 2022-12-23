package uow

import (
	"context"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/user"
)

type FakeUnitOfWorkContext struct {
	UserRepository    *user.FakeUserRepository
	SessionRepository *user.FakeSessionRepository
	LimitsRepository  *user.FakeLimitsRepository
	ChannelRepository *channel.FakeRepository
	WasRollbackCalled bool
	WasCommitCalled   bool
}

func NewFakeUnitOfWorkContext(
	userRepository *user.FakeUserRepository,
	sessionRepository *user.FakeSessionRepository,
	limitsRepository *user.FakeLimitsRepository,
	channelRepository *channel.FakeRepository,
) *FakeUnitOfWorkContext {
	return &FakeUnitOfWorkContext{
		UserRepository:    userRepository,
		SessionRepository: sessionRepository,
		LimitsRepository:  limitsRepository,
		ChannelRepository: channelRepository,
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
