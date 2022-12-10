-- name: CreateUser :one
INSERT INTO "user" (email, identity, password_hash, created_at, activated_at, activation_token) 
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM "user" WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM "user" WHERE email = $1;

-- name: CreateSession :one
INSERT INTO session (token, user_id, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserBySessionToken :one
SELECT "user".* FROM "user" 
JOIN session ON "user".id = session.user_id
WHERE session.token = $1;

-- name: ActivateUser :one
UPDATE "user" 
SET activated_at = @activated_at::timestamp, activation_token = null
WHERE activation_token = @activation_token::text
RETURNING *;

-- name: DeleteSessionByToken :one
DELETE FROM session WHERE token = $1 RETURNING user_id;