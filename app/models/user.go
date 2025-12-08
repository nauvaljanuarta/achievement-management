package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"fullName"`
	RoleID       uuid.UUID `json:"roleId"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type CreateUserRequest struct {
	Username     string    `json:"username" binding:"required"`
	Email        string    `json:"email" binding:"required,email"`
	Password     string    `json:"password" binding:"required"`
	FullName     string    `json:"fullName" binding:"required"`
	RoleID       uuid.UUID `json:"roleId" binding:"required"`
	// Opsional untuk mahasiswa
	StudentID    string    `json:"studentId,omitempty"`
	ProgramStudy string    `json:"programStudy,omitempty"`
	AcademicYear string    `json:"academicYear,omitempty"`
	AdvisorID    uuid.UUID `json:"advisorId,omitempty"`
	// Opsional untuk lecturer
	LecturerID   string    `json:"lecturerId,omitempty"`
	Department   string    `json:"department,omitempty"`
}

// Request untuk update user
type UpdateUserRequest struct {
	FullName     *string    `json:"fullName,omitempty"`
	Email        *string    `json:"email,omitempty"`
	Password     *string    `json:"password,omitempty"`
	IsActive     *bool      `json:"isActive,omitempty"`
	RoleID       *uuid.UUID `json:"roleId,omitempty"`
	ProgramStudy *string    `json:"programStudy,omitempty"`
	AcademicYear *string    `json:"academicYear,omitempty"`
	AdvisorID    *uuid.UUID `json:"advisorId,omitempty"`
	Department   *string    `json:"department,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}
