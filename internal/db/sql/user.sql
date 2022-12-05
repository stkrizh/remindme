-- name: CreateUser :one
INSERT INTO "user" (email, identity, password_hash, created_at, activated_at, activation_token) 
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM "user" WHERE id = $1;

-- name: CreateSession :one
INSERT INTO session (token, user_id, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserBySessionToken :one
SELECT "user".* FROM "user" 
JOIN session ON "user".id = session.user_id
WHERE session.token = $1;