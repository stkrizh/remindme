-- name: CreateChannel :one
INSERT INTO channel (user_id, created_at, type, settings, verification_token, verified_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ReadChanels :many
SELECT * FROM channel WHERE 
    (@all_user_ids::boolean OR user_id = @user_id_equals::bigint)
    AND (@all_types::boolean OR type = @type_equals::text)
ORDER BY id;

-- name: CountChannels :one
SELECT COUNT(id) FROM channel WHERE
    (@all_user_ids::boolean OR user_id = @user_id_equals::bigint)
    AND (@all_types::boolean OR type = @type_equals::text);

-- name: ActivateChannel :one
UPDATE channel 
SET verified_at = @verified_at::timestamp, verification_token = null
WHERE id = $1 AND user_id = $2 AND verification_token = @verification_token::text
RETURNING *;