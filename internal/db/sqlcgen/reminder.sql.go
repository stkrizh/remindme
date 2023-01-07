// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: reminder.sql

package sqlcgen

import (
	"context"
	"database/sql"
	"time"
)

const countReminders = `-- name: CountReminders :one
SELECT COUNT(id) FROM reminder WHERE 
    ($1::boolean OR user_id = $2::bigint)
    AND ($3::boolean OR sent_at >= $4::timestamp)
    AND ($5::boolean OR status = ANY($6::text[]))
`

type CountRemindersParams struct {
	AnyUserID    bool
	UserIDEquals int64
	AnySentAt    bool
	SentAfter    time.Time
	AnyStatus    bool
	StatusIn     []string
}

func (q *Queries) CountReminders(ctx context.Context, arg CountRemindersParams) (int64, error) {
	row := q.db.QueryRow(ctx, countReminders,
		arg.AnyUserID,
		arg.UserIDEquals,
		arg.AnySentAt,
		arg.SentAfter,
		arg.AnyStatus,
		arg.StatusIn,
	)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const createReminder = `-- name: CreateReminder :one
INSERT INTO reminder (user_id, created_at, scheduled_at, sent_at, canceled_at, at, every, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, user_id, created_at, at, status, every, scheduled_at, sent_at, canceled_at
`

type CreateReminderParams struct {
	UserID      int64
	CreatedAt   time.Time
	ScheduledAt sql.NullTime
	SentAt      sql.NullTime
	CanceledAt  sql.NullTime
	At          time.Time
	Every       sql.NullString
	Status      string
}

func (q *Queries) CreateReminder(ctx context.Context, arg CreateReminderParams) (Reminder, error) {
	row := q.db.QueryRow(ctx, createReminder,
		arg.UserID,
		arg.CreatedAt,
		arg.ScheduledAt,
		arg.SentAt,
		arg.CanceledAt,
		arg.At,
		arg.Every,
		arg.Status,
	)
	var i Reminder
	err := row.Scan(
		&i.ID,
		&i.UserID,
		&i.CreatedAt,
		&i.At,
		&i.Status,
		&i.Every,
		&i.ScheduledAt,
		&i.SentAt,
		&i.CanceledAt,
	)
	return i, err
}

type CreateReminderChannelsParams struct {
	ReminderID int64
	ChannelID  int64
}

const readReminders = `-- name: ReadReminders :many
SELECT reminder.id, reminder.user_id, reminder.created_at, reminder.at, reminder.status, reminder.every, reminder.scheduled_at, reminder.sent_at, reminder.canceled_at, array_agg(channel.id ORDER BY channel.id)::bigint[] AS channel_ids FROM reminder 
JOIN reminder_channel ON reminder_channel.reminder_id = reminder.id
JOIN channel ON reminder_channel.channel_id = channel.id
WHERE 
    ($1::boolean OR reminder.user_id = $2::bigint)
    AND ($3::boolean OR reminder.sent_at >= $4::timestamp)
    AND ($5::boolean OR reminder.status = ANY($6::text[]))
GROUP BY reminder.id
ORDER BY 
    CASE WHEN $7::boolean THEN reminder.id ELSE null END,
    CASE WHEN $8::boolean THEN reminder.id ELSE null END DESC,
    CASE WHEN $9::boolean THEN reminder.at ELSE null END,
    CASE WHEN $10::boolean THEN reminder.at ELSE null END DESC,
    id ASC
LIMIT CASE WHEN $12::boolean THEN null ELSE $13::integer END
OFFSET $11::integer
`

type ReadRemindersParams struct {
	AnyUserID     bool
	UserIDEquals  int64
	AnySentAt     bool
	SentAfter     time.Time
	AnyStatus     bool
	StatusIn      []string
	OrderByIDAsc  bool
	OrderByIDDesc bool
	OrderByAtAsc  bool
	OrderByAtDesc bool
	Offset        int32
	AllRows       bool
	Limit         int32
}

type ReadRemindersRow struct {
	ID          int64
	UserID      int64
	CreatedAt   time.Time
	At          time.Time
	Status      string
	Every       sql.NullString
	ScheduledAt sql.NullTime
	SentAt      sql.NullTime
	CanceledAt  sql.NullTime
	ChannelIds  []int64
}

func (q *Queries) ReadReminders(ctx context.Context, arg ReadRemindersParams) ([]ReadRemindersRow, error) {
	rows, err := q.db.Query(ctx, readReminders,
		arg.AnyUserID,
		arg.UserIDEquals,
		arg.AnySentAt,
		arg.SentAfter,
		arg.AnyStatus,
		arg.StatusIn,
		arg.OrderByIDAsc,
		arg.OrderByIDDesc,
		arg.OrderByAtAsc,
		arg.OrderByAtDesc,
		arg.Offset,
		arg.AllRows,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []ReadRemindersRow
	for rows.Next() {
		var i ReadRemindersRow
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.CreatedAt,
			&i.At,
			&i.Status,
			&i.Every,
			&i.ScheduledAt,
			&i.SentAt,
			&i.CanceledAt,
			&i.ChannelIds,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
