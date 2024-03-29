// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0
// source: copyfrom.go

package sqlcgen

import (
	"context"
)

// iteratorForCreateReminderChannels implements pgx.CopyFromSource.
type iteratorForCreateReminderChannels struct {
	rows                 []CreateReminderChannelsParams
	skippedFirstNextCall bool
}

func (r *iteratorForCreateReminderChannels) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}

func (r iteratorForCreateReminderChannels) Values() ([]interface{}, error) {
	return []interface{}{
		r.rows[0].ReminderID,
		r.rows[0].ChannelID,
	}, nil
}

func (r iteratorForCreateReminderChannels) Err() error {
	return nil
}

func (q *Queries) CreateReminderChannels(ctx context.Context, arg []CreateReminderChannelsParams) (int64, error) {
	return q.db.CopyFrom(ctx, []string{"reminder_channel"}, []string{"reminder_id", "channel_id"}, &iteratorForCreateReminderChannels{rows: arg})
}
