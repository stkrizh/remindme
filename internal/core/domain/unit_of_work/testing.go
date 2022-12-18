package uow

import (
	"context"
	"remindme/internal/core/domain/channel"
	"remindme/internal/core/domain/user"
)

type FakeUnitOfWorkContext struct {
	UserRepository    *user.FakeUserRepository
	SessionRepository *user.FakeSessionRepository
	ChannelRepository *channel.FakeRepository
	WasRollbackCalled bool
	WasCommitCalled   bool
}

func NewFakeUnitOfWorkContext(
	userRepository *user.FakeUserRepository,
	sessionRepository *user.FakeSessionRepository,
	channelRepository *channel.FakeRepository,
) *FakeUnitOfWorkContext {
	return &FakeUnitOfWorkContext{
		UserRepository:    userRepository,
		SessionRepository: sessionRepository,
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
			channel.NewFakeRepository(),
		),
	}
}

func (u *FakeUnitOfWork) Begin(ctx context.Context) (Context, error) {
	return u.Context, nil
}
