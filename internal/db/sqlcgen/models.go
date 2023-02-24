// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package sqlcgen

import (
	"database/sql"
	"time"

	"github.com/jackc/pgtype"
)

type Channel struct {
	ID                int64
	UserID            int64
	CreatedAt         time.Time
	IsDefault         bool
	Type              string
	Settings          pgtype.JSONB
	VerificationToken sql.NullString
	VerifiedAt        sql.NullTime
}

type Limit struct {
	ID                       int64
	UserID                   int64
	EmailChannelCount        sql.NullInt32
	TelegramChannelCount     sql.NullInt32
	ActiveReminderCount      sql.NullInt32
	MonthlySentReminderCount sql.NullInt32
	ReminderEveryPerDayCount sql.NullFloat64
}

type Reminder struct {
	ID          int64
	UserID      int64
	CreatedAt   time.Time
	At          time.Time
	Body        string
	Status      string
	Every       sql.NullString
	ScheduledAt sql.NullTime
	SentAt      sql.NullTime
	CanceledAt  sql.NullTime
}

type ReminderChannel struct {
	ID         int64
	ReminderID int64
	ChannelID  int64
}

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
	Timezone        string
	ActivatedAt     sql.NullTime
	ActivationToken sql.NullString
}
