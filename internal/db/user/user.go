package user

import (
	"context"
	"database/sql"
	"errors"
	"remindme/internal/db/sqlcgen"
	c "remindme/internal/domain/common"
	"remindme/internal/domain/user"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

const PG_UNIQUE_CONSTRAINT_ERR_CODE = "23505"
const EMAIL_CONSTRAINT_NAME = "user_email_idx"

type PgxUserRepository struct {
	queries *sqlcgen.Queries
}

func NewPgxRepository(db sqlcgen.DBTX) *PgxUserRepository {
	if db == nil {
		panic("Argument db must not be nil.")
	}
	return &PgxUserRepository{queries: sqlcgen.New(db)}
}

func (r *PgxUserRepository) Create(ctx context.Context, input user.CreateUserInput) (u user.User, err error) {
	dbuser, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		Email:           encodeEmail(input.Email),
		Identity:        encodeIdentity(input.Identity),
		PasswordHash:    encodePasswordHash(input.PasswordHash),
		CreatedAt:       input.CreatedAt,
		ActivatedAt:     encodeOptionalTime(input.ActivatedAt),
		ActivationToken: encodeActivationToken(input.ActivationToken),
	})

	var errEmailUniqueConstraint *pgconn.PgError
	if errors.As(err, &errEmailUniqueConstraint) {
		if errEmailUniqueConstraint.Code == PG_UNIQUE_CONSTRAINT_ERR_CODE &&
			errEmailUniqueConstraint.ConstraintName == EMAIL_CONSTRAINT_NAME {
			return u, &user.EmailAlreadyExistsError{Email: input.Email.Value}
		}
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

func (r *PgxUserRepository) GetByID(ctx context.Context, id user.ID) (u user.User, err error) {
	dbuser, err := r.queries.GetUserByID(ctx, int64(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return u, &user.UserDoesNotExistError{}
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

func encodeEmail(email c.Optional[user.Email]) sql.NullString {
	return sql.NullString{String: string(email.Value), Valid: email.IsPresent}
}

func encodeIdentity(identity c.Optional[user.Identity]) sql.NullString {
	return sql.NullString{String: string(identity.Value), Valid: identity.IsPresent}
}

func encodePasswordHash(ph c.Optional[user.PasswordHash]) sql.NullString {
	return sql.NullString{String: string(ph.Value), Valid: ph.IsPresent}
}

func encodeActivationToken(token c.Optional[user.ActivationToken]) sql.NullString {
	return sql.NullString{String: string(token.Value), Valid: token.IsPresent}
}

func encodeOptionalTime(at c.Optional[time.Time]) sql.NullTime {
	return sql.NullTime{Time: at.Value, Valid: at.IsPresent}
}

func decodeUser(u sqlcgen.User) user.User {
	return user.User{
		ID:              user.ID(u.ID),
		Email:           c.NewOptional(user.Email(u.Email.String), u.Email.Valid),
		Identity:        c.NewOptional(user.Identity(u.Identity.String), u.Identity.Valid),
		PasswordHash:    c.NewOptional(user.PasswordHash(u.PasswordHash.String), u.PasswordHash.Valid),
		CreatedAt:       u.CreatedAt,
		ActivatedAt:     c.NewOptional(u.ActivatedAt.Time, u.ActivatedAt.Valid),
		ActivationToken: c.NewOptional(user.ActivationToken(u.ActivationToken.String), u.ActivationToken.Valid),
		LastLoginAt:     c.NewOptional(u.LastLoginAt.Time, u.LastLoginAt.Valid),
	}
}
