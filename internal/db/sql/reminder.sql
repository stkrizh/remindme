-- name: CreateReminder :one
INSERT INTO reminder (user_id, created_at, scheduled_at, sent_at, canceled_at, at, every, status, body)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;


-- name: CreateReminderChannels :copyfrom
INSERT INTO reminder_channel (reminder_id, channel_id)
VALUES ($1, $2);


-- name: ReadReminders :many
SELECT reminder.*, array_agg(channel.id ORDER BY channel.id)::bigint[] AS channel_ids FROM reminder 
JOIN reminder_channel ON reminder_channel.reminder_id = reminder.id
JOIN channel ON reminder_channel.channel_id = channel.id
WHERE 
    (@any_user_id::boolean OR reminder.user_id = @user_id_equals::bigint)
    AND (@any_sent_at::boolean OR reminder.sent_at >= @sent_after::timestamp)
    AND (@any_status::boolean OR reminder.status = ANY(@status_in::text[]))
GROUP BY reminder.id
ORDER BY 
    CASE WHEN @order_by_id_asc::boolean THEN reminder.id ELSE null END,
    CASE WHEN @order_by_id_desc::boolean THEN reminder.id ELSE null END DESC,
    CASE WHEN @order_by_at_asc::boolean THEN reminder.at ELSE null END,
    CASE WHEN @order_by_at_desc::boolean THEN reminder.at ELSE null END DESC,
    id ASC
LIMIT CASE WHEN @all_rows::boolean THEN null ELSE @limit_::integer END
OFFSET @offset_::integer;


-- name: CountReminders :one
SELECT COUNT(id) FROM reminder WHERE 
    (@any_user_id::boolean OR user_id = @user_id_equals::bigint)
    AND (@any_sent_at::boolean OR sent_at >= @sent_after::timestamp)
    AND (@any_status::boolean OR status = ANY(@status_in::text[]));


-- name: LockReminder :exec
SELECT pg_advisory_xact_lock(@reminder_id::bigint);


-- name: GetReminderByID :one
SELECT reminder.*, array_agg(channel.id ORDER BY channel.id)::bigint[] AS channel_ids 
FROM reminder
JOIN reminder_channel ON reminder_channel.reminder_id = reminder.id
JOIN channel ON reminder_channel.channel_id = channel.id
WHERE reminder.id = $1
GROUP BY reminder.id;


-- name: UpdateReminder :one
UPDATE reminder 
SET 
    at = CASE WHEN @do_at_update::boolean THEN @at
        ELSE at END,
    body = CASE WHEN @do_body_update::boolean THEN @body
        ELSE body END,
    every = CASE WHEN @do_every_update::boolean THEN @every
        ELSE every END,
    status = CASE WHEN @do_status_update::boolean THEN @status
        ELSE status END,
    scheduled_at = CASE WHEN @do_scheduled_at_update::boolean THEN @scheduled_at
        ELSE scheduled_at END,
    sent_at = CASE WHEN @do_sent_at_update::boolean THEN @sent_at
        ELSE sent_at END,
    canceled_at = CASE WHEN @do_canceled_at_update::boolean THEN @canceled_at
        ELSE canceled_at END
WHERE id = $1
RETURNING *;