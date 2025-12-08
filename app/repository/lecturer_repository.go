package repository

import (
	"database/sql"
	"errors"

	"achievement-backend/app/models"
)

type LecturerRepository interface {
	GetByUserID(userID string) (*models.Lecturer, error)
	Create(lecturer models.Lecturer) (string, error)
	GetAll() ([]models.Lecturer, error)
}

type lecturerRepo struct {
	DB *sql.DB
}

func NewLecturerRepository(db *sql.DB) LecturerRepository {
	return &lecturerRepo{DB: db}
}

func (r *lecturerRepo) GetByUserID(userID string) (*models.Lecturer, error) {
	var l models.Lecturer
	err := r.DB.QueryRow(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers WHERE user_id=$1
	`, userID).Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *lecturerRepo) Create(lecturer models.Lecturer) (string, error) {
	var id string
	err := r.DB.QueryRow(`
		INSERT INTO lecturers (user_id, lecturer_id, department, created_at)
		VALUES ($1, $2, $3, NOW()) RETURNING id
	`, lecturer.UserID, lecturer.LecturerID, lecturer.Department).Scan(&id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *lecturerRepo) GetAll() ([]models.Lecturer, error) {
	rows, err := r.DB.Query(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lecturers []models.Lecturer
	for rows.Next() {
		var l models.Lecturer
		if err := rows.Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt); err != nil {
			return nil, err
		}
		lecturers = append(lecturers, l)
	}
	return lecturers, nil
}
