package repository

import (
	"database/sql"
	"errors"

	"achievement-backend/app/models"
)

type StudentRepository interface {
	GetByUserID(userID string) (*models.Student, error)
	Create(student models.Student) (string, error)
	GetAll() ([]models.Student, error)
	GetAllByAdvisorID(advisorID string) ([]models.Student, error)
}

type studentRepo struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepo{DB: db}
}

func (r *studentRepo) GetByUserID(userID string) (*models.Student, error) {
	var s models.Student
	err := r.DB.QueryRow(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students WHERE user_id=$1
	`, userID).Scan(&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepo) Create(student models.Student) (string, error) {
	var id string
	err := r.DB.QueryRow(`
		INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW()) RETURNING id
	`, student.UserID, student.StudentID, student.ProgramStudy, student.AcademicYear, student.AdvisorID).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *studentRepo) GetAll() ([]models.Student, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	return students, nil
}

func (r *studentRepo) GetAllByAdvisorID(advisorID string) ([]models.Student, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, advisor_id, created_at
		FROM students WHERE advisor_id=$1 ORDER BY created_at DESC
	`, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt); err != nil {
			return nil, err
		}
		students = append(students, s)
	}
	return students, nil
}
