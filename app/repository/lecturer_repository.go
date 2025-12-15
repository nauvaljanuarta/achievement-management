package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"achievement-backend/app/models"
	"github.com/google/uuid"
)

type LecturerRepository interface {
	GetByID(id uuid.UUID) (*models.Lecturer, error)
	GetByUserID(userID uuid.UUID) (*models.Lecturer, error)
	GetByLecturerID(lecturerID string) (*models.Lecturer, error)
	Create(lecturer models.Lecturer) (uuid.UUID, error)
	Update(id uuid.UUID, req *models.UpdateLecturerRequest) error
	
	GetAll(page, limit int) ([]models.Lecturer, int, error)
	GetTotalCount() (int, error)
	GetWithUserDetails(page, limit int) ([]models.LecturerResponse, int, error)
	
	GetAdviseesCount(lecturerID uuid.UUID) (int, error)
	GetAdvisees(lecturerID uuid.UUID, page, limit int) ([]models.Student, int, error)
	
	SearchByName(name string, page, limit int) ([]models.LecturerResponse, int, error)
	GetByDepartment(department string, page, limit int) ([]models.Lecturer, int, error)
}

type lecturerRepo struct {
	DB *sql.DB
}

func NewLecturerRepository(db *sql.DB) LecturerRepository {
	return &lecturerRepo{DB: db}
}

func (r *lecturerRepo) GetByID(id uuid.UUID) (*models.Lecturer, error) {
	var l models.Lecturer
	err := r.DB.QueryRow(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers 
		WHERE id=$1
	`, id).Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *lecturerRepo) GetByUserID(userID uuid.UUID) (*models.Lecturer, error) {
	var l models.Lecturer
	err := r.DB.QueryRow(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers 
		WHERE user_id=$1
	`, userID).Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *lecturerRepo) GetByLecturerID(lecturerID string) (*models.Lecturer, error) {
	var l models.Lecturer
	err := r.DB.QueryRow(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers 
		WHERE lecturer_id=$1
	`, lecturerID).Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &l, nil
}

func (r *lecturerRepo) Create(lecturer models.Lecturer) (uuid.UUID, error) {
	_, err := r.DB.Exec(`
		INSERT INTO lecturers (id, user_id, lecturer_id, department, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, 
		lecturer.ID,         
		lecturer.UserID,     
		lecturer.LecturerID, 
		lecturer.Department, 
	)
	
	if err != nil {
		return uuid.Nil, err
	}
	return lecturer.ID, nil
}

func (r *lecturerRepo) Update(id uuid.UUID, req *models.UpdateLecturerRequest) error {
	query := `UPDATE lecturers SET`
	params := []interface{}{}
	paramIndex := 1
	
	if req.LecturerID != nil {
		if paramIndex > 1 {
			query += ","
		}
		query += fmt.Sprintf(` lecturer_id=$%d`, paramIndex)
		params = append(params, *req.LecturerID)
		paramIndex++
	}
	
	if req.Department != nil {
		if paramIndex > 1 {
			query += ","
		}
		query += fmt.Sprintf(` department=$%d`, paramIndex)
		params = append(params, *req.Department)
		paramIndex++
	}
	
	if paramIndex == 1 {
		return errors.New("no fields to update")
	}
	
	query += fmt.Sprintf(` WHERE id=$%d`, paramIndex)
	params = append(params, id)
	
	_, err := r.DB.Exec(query, params...)
	return err
}

func (r *lecturerRepo) GetAll(page, limit int) ([]models.Lecturer, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lecturers []models.Lecturer
	for rows.Next() {
		var l models.Lecturer
		if err := rows.Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		lecturers = append(lecturers, l)
	}

	// Get total count
	total, err := r.GetTotalCount()
	if err != nil {
		return nil, 0, err
	}

	return lecturers, total, nil
}

func (r *lecturerRepo) GetTotalCount() (int, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM lecturers`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *lecturerRepo) GetWithUserDetails(page, limit int) ([]models.LecturerResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT l.id, l.user_id, u.full_name, u.email, u.username,
		       l.lecturer_id, l.department, l.created_at
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE u.is_active = true
		ORDER BY u.full_name
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lecturers []models.LecturerResponse
	for rows.Next() {
		var l models.LecturerResponse
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.FullName, &l.Email, &l.Username,
			&l.LecturerID, &l.Department, &l.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		lecturers = append(lecturers, l)
	}

	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE u.is_active = true
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return lecturers, total, nil
}

func (r *lecturerRepo) GetAdviseesCount(lecturerID uuid.UUID) (int, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM students 
		WHERE advisor_id=$1
	`, lecturerID).Scan(&count)
	
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *lecturerRepo) GetAdvisees(lecturerID uuid.UUID, page, limit int) ([]models.Student, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT s.id, s.user_id, s.student_id, s.program_study, 
		       s.academic_year, s.advisor_id, s.created_at
		FROM students s
		WHERE s.advisor_id=$1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`, lecturerID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy,
			&s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}

	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM students
		WHERE advisor_id=$1
	`, lecturerID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *lecturerRepo) SearchByName(name string, page, limit int) ([]models.LecturerResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	searchPattern := "%" + name + "%"
	rows, err := r.DB.Query(`
		SELECT l.id, l.user_id, u.full_name, u.email, 
		       l.lecturer_id, l.department, l.created_at
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE u.is_active = true AND u.full_name ILIKE $1
		ORDER BY u.full_name
		LIMIT $2 OFFSET $3
	`, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lecturers []models.LecturerResponse
	for rows.Next() {
		var l models.LecturerResponse
		if err := rows.Scan(
			&l.ID, &l.UserID, &l.FullName, &l.Email,
			&l.LecturerID, &l.Department, &l.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		lecturers = append(lecturers, l)
	}

	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM lecturers l
		JOIN users u ON l.user_id = u.id
		WHERE u.is_active = true AND u.full_name ILIKE $1
	`, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return lecturers, total, nil
}

func (r *lecturerRepo) GetByDepartment(department string, page, limit int) ([]models.Lecturer, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, user_id, lecturer_id, department, created_at
		FROM lecturers 
		WHERE department=$1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, department, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lecturers []models.Lecturer
	for rows.Next() {
		var l models.Lecturer
		if err := rows.Scan(&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		lecturers = append(lecturers, l)
	}

	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM lecturers
		WHERE department=$1
	`, department).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return lecturers, total, nil
}