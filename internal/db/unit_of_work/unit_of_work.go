package uow

import (
	"context"
	"remindme/internal/core/domain/channel"
	uow "remindme/internal/core/domain/unit_of_work"
	"remindme/internal/core/domain/user"
	dbchannel "remindme/internal/db/channel"
	dbuser "remindme/internal/db/user"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type pgxUnitOfWorkContext struct {
	tx pgx.Tx
}

func newPgxUnitOfWorkContext(tx pgx.Tx) *pgxUnitOfWorkContext {
	return &pgxUnitOfWorkContext{
		tx: tx,
	}
}

func (c *pgxUnitOfWorkContext) Commit(ctx context.Context) error {
	return c.tx.Commit(ctx)
}

func (c *pgxUnitOfWorkContext) Rollback(ctx context.Context) error {
	return c.tx.Rollback(ctx)
}

func (c *pgxUnitOfWorkContext) Users() user.UserRepository {
	return dbuser.NewPgxRepository(c.tx)
}

func (c *pgxUnitOfWorkContext) Sessions() user.SessionRepository {
	return dbuser.NewPgxSessionRepository(c.tx)
}

func (c *pgxUnitOfWorkContext) Limits() user.LimitsRepository {
	return dbuser.NewPgxLimitsRepository(c.tx)
}

func (c *pgxUnitOfWorkContext) Channels() channel.Repository {
	return dbchannel.NewPgxChannelRepository(c.tx)
}

type PgxUnitOfWork struct {
	db *pgxpool.Pool
}

func NewPgxUnitOfWork(db *pgxpool.Pool) *PgxUnitOfWork {
	if db == nil {
		panic("Argument db must not be nil.")
	}
	return &PgxUnitOfWork{db: db}
}

func (u *PgxUnitOfWork) Begin(ctx context.Context) (uow.Context, error) {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return newPgxUnitOfWorkContext(tx), nil
}
