-- name: CreateApplication :one
INSERT INTO applications(id, created_at, updated_at, cover_note, job_id, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetApplicationsByUserID :many
SELECT *
FROM applications
WHERE user_id = $1;

-- name: GetApplicationsByJobID :many
SELECT *
FROM applications
WHERE job_id = $1;