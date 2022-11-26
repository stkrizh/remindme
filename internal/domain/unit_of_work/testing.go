package uow

import (
	"context"
	"remindme/internal/domain/user"
)

type FakeUnitOfWorkContext struct {
	UserRepository    user.Repository
	WasRollbackCalled bool
	WasCommitCalled   bool
}

func NewFakeUnitOfWorkContext(userRepository user.Repository) *FakeUnitOfWorkContext {
	return &FakeUnitOfWorkContext{UserRepository: userRepository}
}

func (c *FakeUnitOfWorkContext) Rollback(ctx context.Context) error {
	c.WasRollbackCalled = true
	return nil
}

func (c *FakeUnitOfWorkContext) Commit(ctx context.Context) error {
	c.WasCommitCalled = true
	return nil
}

func (c *FakeUnitOfWorkContext) Users() user.Repository {
	return c.UserRepository
}

type FakeUnitOfWork struct {
	Context *FakeUnitOfWorkContext
}

func NewFakeUnitOfWork(context *FakeUnitOfWorkContext) *FakeUnitOfWork {
	return &FakeUnitOfWork{Context: context}
}

func (u *FakeUnitOfWork) Begin(ctx context.Context) (Context, error) {
	return u.Context, nil
}
