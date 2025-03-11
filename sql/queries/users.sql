-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, username, password_hash)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, username;

-- name: GetUsers :many
SELECT id, created_at, updated_at, username
FROM users;

-- name: GetUserByUsername :one
SELECT id, created_at, updated_at, username
FROM users
WHERE username = $1;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetPasswordHashByUsername :one
SELECT password_hash
FROM users
WHERE username = $1;