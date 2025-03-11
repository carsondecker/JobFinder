-- name: CreateJob :one
INSERT INTO jobs(id, created_at, updated_at, title, description, city, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: GetJobs :many
SELECT *
FROM jobs;

-- name: GetJobByID :one
SELECT *
FROM jobs
WHERE id = $1;

-- name: GetJobsByTitle :many
SELECT *
FROM jobs
WHERE title ILIKE '%' || CAST($1 AS TEXT) || '%';

-- name: DeleteJob :exec
DELETE FROM jobs
WHERE id = $1;