package uow

import (
	"context"
	"remindme/internal/domain/user"
)

type FakeUnitOfWorkContext struct {
	UserRepository    user.UserRepository
	SessionRepository user.SessionRepository
	WasRollbackCalled bool
	WasCommitCalled   bool
}

func NewFakeUnitOfWorkContext(
	userRepository user.UserRepository,
	sessionRepository user.SessionRepository,
) *FakeUnitOfWorkContext {
	return &FakeUnitOfWorkContext{
		UserRepository:    userRepository,
		SessionRepository: sessionRepository,
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

type FakeUnitOfWork struct {
	Context *FakeUnitOfWorkContext
}

func NewFakeUnitOfWork() *FakeUnitOfWork {
	userRepository := user.NewFakeUserRepository()
	return &FakeUnitOfWork{
		Context: NewFakeUnitOfWorkContext(
			userRepository,
			user.NewFakeSessionRepository(userRepository),
		),
	}
}

func (u *FakeUnitOfWork) Begin(ctx context.Context) (Context, error) {
	return u.Context, nil
}
