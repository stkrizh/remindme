-- name: CreateChannel :one
INSERT INTO channel (user_id, created_at, settings, is_verified)
VALUES ($1, $2, $3, $4)
RETURNING *;