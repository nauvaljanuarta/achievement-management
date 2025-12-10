package models

import (
	"time"

	"github.com/google/uuid"
)

type Student struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	UserID       uuid.UUID  `json:"userId" db:"user_id"`
	StudentID    string     `json:"studentId" db:"student_id"`
	ProgramStudy string     `json:"programStudy" db:"program_study"`
	AcademicYear string     `json:"academicYear" db:"academic_year"`
	AdvisorID    *uuid.UUID `json:"advisorId" db:"advisor_id"`
	CreatedAt    time.Time  `json:"createdAt" db:"created_at"`
}

type CreateStudentProfileRequest struct {
	StudentID    string  `json:"studentId" binding:"required"`
	ProgramStudy string  `json:"programStudy" binding:"required"`
	AcademicYear string  `json:"academicYear" binding:"required"`
	AdvisorID    *string `json:"advisorId,omitempty" binding:"omitempty,uuid"`
}

type UpdateStudentRequest struct {
	ProgramStudy *string `json:"programStudy,omitempty"`
	AcademicYear *string `json:"academicYear,omitempty"`
	AdvisorID    *string `json:"advisorId,omitempty" binding:"omitempty,uuid"`
}

type StudentResponse struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"userId"`
	FullName     string     `json:"fullName"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	StudentID    string     `json:"studentId"`
	ProgramStudy string     `json:"programStudy"`
	AcademicYear string     `json:"academicYear"`
	AdvisorID    *uuid.UUID `json:"advisorId,omitempty"`
	AdvisorName  string     `json:"advisorName,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}