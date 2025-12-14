package models

import "github.com/google/uuid"

type MetaInfo struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Total  int    `json:"total"`
	Pages  int    `json:"pages"`
	SortBy string `json:"sortBy"`
	Order  string `json:"order"`
	Search string `json:"search"`
}

type StudentDetail struct {
	StudentID    string    `json:"studentId"`
	ProgramStudy string    `json:"programStudy"`
	AcademicYear string    `json:"academicYear"`
	AdvisorID    uuid.UUID `json:"advisorId"`
}

type LecturerDetail struct {
	LecturerID string `json:"lecturerId"`
	Department string `json:"department"`
}

type UserProfileResponse struct {
	ID             uuid.UUID      `json:"id"`
	Username       string         `json:"username"`
	FullName       string         `json:"fullName"`
	Email          string         `json:"email"`
	Role           string         `json:"role"`
	IsActive       bool           `json:"isActive"`
	CreatedAt      string         `json:"createdAt"`
	UpdatedAt      string         `json:"updatedAt"`
	StudentDetail  *StudentDetail  `json:"studentDetail,omitempty"`
	LecturerDetail *LecturerDetail `json:"lecturerDetail,omitempty"`
}

type UserResponse struct {
	Data []UserProfileResponse `json:"data"`
	Meta MetaInfo              `json:"meta"`
}


