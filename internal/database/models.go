// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0

package database

import (
	"time"

	"github.com/google/uuid"
)

type Application struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	CoverNote string
	JobID     uuid.UUID
	UserID    uuid.UUID
}

type Job struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       string
	Description string
	City        string
	UserID      uuid.UUID
}

type User struct {
	ID           uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Username     string
	PasswordHash string
}
