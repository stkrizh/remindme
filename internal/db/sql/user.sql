-- name: CreateUser :one
INSERT INTO "user" (email, identity, password_hash, created_at, activated_at, activation_token) 
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM "user" WHERE id = $1;