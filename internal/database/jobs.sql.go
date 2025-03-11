// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: jobs.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createJob = `-- name: CreateJob :one
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
RETURNING id, created_at, updated_at, title, description, city, user_id
`

type CreateJobParams struct {
	Title       string
	Description string
	City        string
	UserID      uuid.UUID
}

func (q *Queries) CreateJob(ctx context.Context, arg CreateJobParams) (Job, error) {
	row := q.db.QueryRowContext(ctx, createJob,
		arg.Title,
		arg.Description,
		arg.City,
		arg.UserID,
	)
	var i Job
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Title,
		&i.Description,
		&i.City,
		&i.UserID,
	)
	return i, err
}

const deleteJob = `-- name: DeleteJob :exec
DELETE FROM jobs
WHERE id = $1
`

func (q *Queries) DeleteJob(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteJob, id)
	return err
}

const getJobByID = `-- name: GetJobByID :one
SELECT id, created_at, updated_at, title, description, city, user_id
FROM jobs
WHERE id = $1
`

func (q *Queries) GetJobByID(ctx context.Context, id uuid.UUID) (Job, error) {
	row := q.db.QueryRowContext(ctx, getJobByID, id)
	var i Job
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Title,
		&i.Description,
		&i.City,
		&i.UserID,
	)
	return i, err
}

const getJobs = `-- name: GetJobs :many
SELECT id, created_at, updated_at, title, description, city, user_id
FROM jobs
`

func (q *Queries) GetJobs(ctx context.Context) ([]Job, error) {
	rows, err := q.db.QueryContext(ctx, getJobs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Job
	for rows.Next() {
		var i Job
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Title,
			&i.Description,
			&i.City,
			&i.UserID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getJobsByTitle = `-- name: GetJobsByTitle :many
SELECT id, created_at, updated_at, title, description, city, user_id
FROM jobs
WHERE title ILIKE '%' || CAST($1 AS TEXT) || '%'
`

func (q *Queries) GetJobsByTitle(ctx context.Context, dollar_1 string) ([]Job, error) {
	rows, err := q.db.QueryContext(ctx, getJobsByTitle, dollar_1)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Job
	for rows.Next() {
		var i Job
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Title,
			&i.Description,
			&i.City,
			&i.UserID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
