package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FullName     string    `json:"fullName" db:"full_name"`
	RoleID       uuid.UUID `json:"roleId" db:"role_id"`
	IsActive     bool      `json:"isActive" db:"is_active"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

type CreateUserRequest struct {
	Username  string `json:"username" binding:"required,alphanum,min=3"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FullName  string `json:"fullName" binding:"required"`
	RoleID    string `json:"roleId" binding:"required,uuid"`
	
	StudentID    *string `json:"studentId,omitempty"`
	ProgramStudy *string `json:"programStudy,omitempty"`
	AcademicYear *string `json:"academicYear,omitempty"`
	AdvisorID    *string `json:"advisorId,omitempty" binding:"omitempty,uuid"`
	
	LecturerID   *string `json:"lecturerId,omitempty"`
	Department   *string `json:"department,omitempty"`
}

type UpdateUserRequest struct {
	FullName *string `json:"fullName,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty" binding:"omitempty,min=6"`
	IsActive *bool   `json:"isActive,omitempty"`
	RoleID   *string `json:"roleId,omitempty" binding:"omitempty,uuid"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	User         User     `json:"user"`
	Permissions  []string `json:"permissions"` 
}

type JWTClaims struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	RoleID      string `json:"role_id"`
	Permissions []string `json:"permissions"` 
	jwt.RegisteredClaims
}