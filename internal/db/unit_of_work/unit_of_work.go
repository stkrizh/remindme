package uow

import (
	"context"
	"remindme/internal/db/sqlcgen"
	dbuser "remindme/internal/db/user"
	uow "remindme/internal/domain/unit_of_work"
	"remindme/internal/domain/user"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type pgxUnitOfWorkContext struct {
	tx      pgx.Tx
	queries *sqlcgen.Queries
}

func newPgxUnitOfWorkContext(tx pgx.Tx) *pgxUnitOfWorkContext {
	return &pgxUnitOfWorkContext{
		tx:      tx,
		queries: sqlcgen.New(tx),
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
