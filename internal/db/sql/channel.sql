-- name: CreateChannel :one
INSERT INTO channel (user_id, created_at, is_default, type, settings, verification_token, verified_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ReadChanels :many
SELECT * FROM channel WHERE 
    (@all_channel_ids::boolean OR id = ANY(@id_in::bigint[])) 
    AND (@all_user_ids::boolean OR user_id = @user_id_equals::bigint)
    AND (@all_types::boolean OR type = @type_equals::text)
    AND (@all_is_default::boolean OR is_default = @is_default_equals::boolean)
ORDER BY 
    CASE WHEN @order_by_id_asc::boolean THEN channel.id ELSE null END ASC,
    CASE WHEN @order_by_id_desc::boolean THEN channel.id ELSE null END DESC,
    id ASC
LIMIT CASE WHEN @all_rows::boolean THEN null ELSE @limit_::integer END;

-- name: GetChannelByID :one
SELECT * FROM channel WHERE id = $1;

-- name: CountChannels :one
SELECT COUNT(id) FROM channel WHERE
    (@all_channel_ids::boolean OR id = ANY(@id_in::bigint[])) 
    AND (@all_user_ids::boolean OR user_id = @user_id_equals::bigint)
    AND (@all_types::boolean OR type = @type_equals::text)
    AND (@all_is_default::boolean OR is_default = @is_default_equals::boolean);

-- name: UpdateChannel :one
UPDATE channel 
SET 
    verification_token = CASE WHEN @do_verification_token_update::boolean THEN @verification_token
        ELSE verification_token END,
    verified_at = CASE WHEN @do_verified_at_update::boolean THEN @verified_at
        ELSE verified_at END,
    settings = CASE WHEN @do_settings_update::boolean THEN @settings
        ELSE settings END
WHERE id = $1
RETURNING *;