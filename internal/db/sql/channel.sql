-- name: CreateChannel :one
INSERT INTO channel (user_id, created_at, settings, verification_token, verified_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;