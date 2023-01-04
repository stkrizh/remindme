-- name: CreateReminder :one
INSERT INTO reminder (user_id, created_at, scheduled_at, sent_at, canceled_at, at, every, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;


-- name: CreateReminderChannels :copyfrom
INSERT INTO reminder_channel (reminder_id, channel_id)
VALUES ($1, $2);


-- name: ReadReminders :many
SELECT * FROM reminder WHERE 
    (@any_user_id::boolean OR user_id = @user_id_equals::bigint)
    AND (@any_sent_at::boolean OR sent_at >= @sent_after::timestamp)
    AND (@any_status::boolean OR status = ANY(@status_in::text[]))
ORDER BY 
    CASE WHEN @order_by_id_asc::boolean THEN id ELSE null END,
    CASE WHEN @order_by_id_desc::boolean THEN id ELSE null END DESC,
    CASE WHEN @order_by_at_asc::boolean THEN at ELSE null END,
    CASE WHEN @order_by_at_desc::boolean THEN at ELSE null END DESC,
    id ASC;


-- name: CountReminders :one
SELECT COUNT(id) FROM reminder WHERE 
    (@any_user_id::boolean OR user_id = @user_id_equals::bigint)
    AND (@any_sent_at::boolean OR sent_at >= @sent_after::timestamp)
    AND (@any_status::boolean OR status = ANY(@status_in::text[]));