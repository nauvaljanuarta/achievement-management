package models

import (
	"time"
	"github.com/google/uuid"
)

type Lecturer struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"userId" db:"user_id"`
	LecturerID string    `json:"lecturerId" db:"lecturer_id"`
	Department string    `json:"department" db:"department"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}

type CreateLecturerProfileRequest struct {
	LecturerID string `json:"lecturerId" binding:"required"`
	Department string `json:"department" binding:"required"`
}

type UpdateLecturerRequest struct {
	LecturerID *string `json:"lecturerId,omitempty"`
	Department *string `json:"department,omitempty"`
}

type LecturerResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"userId"`
	FullName   string    `json:"fullName"`
	Username   string    `json:"username"`
	Email      string    `json:"email"`
	LecturerID string    `json:"lecturerId"`
	Department string    `json:"department"`
	CreatedAt  time.Time `json:"createdAt"`
}