-- name: CreateChannel :one
INSERT INTO channel (user_id, created_at, type, settings, verification_token, verified_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ReadChanels :many
SELECT * FROM channel WHERE 
    (@all_user_ids::boolean OR user_id = @user_id_equals::bigint)
    AND (@all_types::boolean OR type = @type_equals::text)
ORDER BY id;

-- name: GetChannelByID :one
SELECT * FROM channel WHERE id = $1;

-- name: CountChannels :one
SELECT COUNT(id) FROM channel WHERE
    (@all_user_ids::boolean OR user_id = @user_id_equals::bigint)
    AND (@all_types::boolean OR type = @type_equals::text);

-- name: UpdateChannel :one
UPDATE channel 
SET 
    verification_token = CASE WHEN @do_verification_token_update::boolean THEN @verification_token
        ELSE verification_token END,
    verified_at = CASE WHEN @do_verified_at_update::boolean THEN @verified_at
        ELSE verified_at END
WHERE id = $1
RETURNING *;