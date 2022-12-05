// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package sqlcgen

import (
	"database/sql"
	"time"
)

type Session struct {
	ID        int64
	Token     string
	UserID    int64
	CreatedAt time.Time
}

type User struct {
	ID              int64
	Email           sql.NullString
	Identity        sql.NullString
	PasswordHash    sql.NullString
	CreatedAt       time.Time
	ActivatedAt     sql.NullTime
	ActivationToken sql.NullString
	LastLoginAt     sql.NullTime
}
