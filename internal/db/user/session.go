package user

import (
	"context"
	"errors"
	"remindme/internal/db/sqlcgen"
	e "remindme/internal/domain/errors"
	"remindme/internal/domain/user"

	"github.com/jackc/pgx/v4"
)

type PgxSessionRepository struct {
	queries *sqlcgen.Queries
}

func NewPgxSessionRepository(db sqlcgen.DBTX) *PgxSessionRepository {
	if db == nil {
		panic(e.NewNilArgumentError("db"))
	}
	return &PgxSessionRepository{queries: sqlcgen.New(db)}
}

func (r *PgxSessionRepository) Create(ctx context.Context, input user.CreateSessionInput) error {
	_, err := r.queries.CreateSession(ctx, sqlcgen.CreateSessionParams{
		Token:     string(input.Token),
		UserID:    int64(input.UserID),
		CreatedAt: input.CreatedAt,
	})
	return err
}

func (r *PgxSessionRepository) GetUserByToken(ctx context.Context, token user.SessionToken) (u user.User, err error) {
	dbuser, err := r.queries.GetUserBySessionToken(ctx, string(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return u, user.ErrUserDoesNotExist
	}
	if err != nil {
		return u, err
	}
	u = decodeUser(dbuser)
	err = u.Validate()
	if err != nil {
		return u, err
	}
	return u, nil
}

func (r *PgxSessionRepository) Delete(ctx context.Context, token user.SessionToken) (userID user.ID, err error) {
	rawUserID, err := r.queries.DeleteSessionByToken(ctx, string(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return userID, user.ErrSessionDoesNotExist
	}
	return user.ID(rawUserID), nil
}
