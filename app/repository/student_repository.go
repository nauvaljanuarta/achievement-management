package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"achievement-backend/app/models"
	"github.com/google/uuid"
)

type StudentRepository interface {
	GetByID(id uuid.UUID) (*models.Student, error)
	GetByUserID(userID uuid.UUID) (*models.Student, error)
	GetByStudentID(studentID string) (*models.Student, error)
	Create(student models.Student) (uuid.UUID, error)
	Update(id uuid.UUID, req *models.UpdateStudentRequest) error
	UpdateAdvisor(studentID, advisorID uuid.UUID) error
	RemoveAdvisor(studentID uuid.UUID) error
	GetAll(page, limit int) ([]models.Student, int, error)
	GetTotalCount() (int, error)
	
	GetWithUserDetails(page, limit int) ([]models.StudentResponse, int, error)
	GetWithAdvisorDetails(page, limit int) ([]models.StudentResponse, int, error)
	
	// Get by advisor with pagination
	GetAllByAdvisorID(advisorID uuid.UUID, page, limit int) ([]models.Student, int, error)
	GetAdvisorless(page, limit int) ([]models.Student, int, error) 
	
	SearchByName(name string, page, limit int) ([]models.StudentResponse, int, error)
	GetByProgramStudy(programStudy string, page, limit int) ([]models.Student, int, error)
	GetByAcademicYear(academicYear string, page, limit int) ([]models.Student, int, error)
	
	GetStudentsCountByAdvisor(advisorID uuid.UUID) (int, error)
	GetStudentsCountByProgramStudy() (map[string]int, error)
}

type studentRepo struct {
	DB *sql.DB
}

func NewStudentRepository(db *sql.DB) StudentRepository {
	return &studentRepo{DB: db}
}

func (r *studentRepo) GetByID(id uuid.UUID) (*models.Student, error) {
	var s models.Student
	err := r.DB.QueryRow(`
		SELECT id, user_id, student_id, program_study, academic_year, 
		       advisor_id, created_at
		FROM students 
		WHERE id=$1
	`, id).Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, 
		&s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepo) GetByUserID(userID uuid.UUID) (*models.Student, error) {
	var s models.Student
	err := r.DB.QueryRow(`
		SELECT id, user_id, student_id, program_study, academic_year, 
		       advisor_id, created_at
		FROM students 
		WHERE user_id=$1
	`, userID).Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, 
		&s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepo) GetByStudentID(studentID string) (*models.Student, error) {
	var s models.Student
	err := r.DB.QueryRow(`
		SELECT id, user_id, student_id, program_study, academic_year, 
		       advisor_id, created_at
		FROM students 
		WHERE student_id=$1
	`, studentID).Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, 
		&s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *studentRepo) Create(student models.Student) (uuid.UUID, error) {
	_, err := r.DB.Exec(`
		INSERT INTO students (id, user_id, student_id, program_study, academic_year, 
		                     advisor_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, 
		student.ID,           // 1
		student.UserID,       // 2
		student.StudentID,    // 3
		student.ProgramStudy, 
		student.AcademicYear, 
		student.AdvisorID,    
	)
	
	if err != nil {
		return uuid.Nil, err
	}
	return student.ID, nil
}

func (r *studentRepo) Update(id uuid.UUID, req *models.UpdateStudentRequest) error {
	query := `UPDATE students SET`
	params := []interface{}{}
	paramIndex := 1
	
	if req.ProgramStudy != nil {
		query += fmt.Sprintf(` program_study=$%d`, paramIndex)
		params = append(params, *req.ProgramStudy)
		paramIndex++
	}
	
	if req.AcademicYear != nil {
		if paramIndex > 1 {
			query += ","
		}
		query += fmt.Sprintf(` academic_year=$%d`, paramIndex)
		params = append(params, *req.AcademicYear)
		paramIndex++
	}
	
	if req.AdvisorID != nil {
		if paramIndex > 1 {
			query += ","
		}
		query += fmt.Sprintf(` advisor_id=$%d`, paramIndex)
		params = append(params, *req.AdvisorID)
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

func (r *studentRepo) UpdateAdvisor(studentID, advisorID uuid.UUID) error {
	_, err := r.DB.Exec(`
		UPDATE students 
		SET advisor_id=$1
		WHERE id=$2
	`, advisorID, studentID)
	return err
}

func (r *studentRepo) RemoveAdvisor(studentID uuid.UUID) error {
	_, err := r.DB.Exec(`
		UPDATE students 
		SET advisor_id=NULL
		WHERE id=$1
	`, studentID)
	return err
}

func (r *studentRepo) GetAll(page, limit int) ([]models.Student, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, 
		       advisor_id, created_at
		FROM students 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
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

	total, err := r.GetTotalCount()
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepo) GetTotalCount() (int, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM students`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *studentRepo) GetWithUserDetails(page, limit int) ([]models.StudentResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT s.id, s.user_id, u.full_name, u.email, u.username,
		       s.student_id, s.program_study, s.academic_year, 
		       s.advisor_id, s.created_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE u.is_active = true
		ORDER BY u.full_name
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []models.StudentResponse
	for rows.Next() {
		var s models.StudentResponse
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.FullName, &s.Email, &s.Username,
			&s.StudentID, &s.ProgramStudy, &s.AcademicYear,
			&s.AdvisorID, &s.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}

	// Get total count
	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE u.is_active = true
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepo) GetWithAdvisorDetails(page, limit int) ([]models.StudentResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT s.id, s.user_id, u.full_name AS student_name,
		       s.student_id, s.program_study, s.academic_year, 
		       s.advisor_id, l.lecturer_id, lu.full_name AS advisor_name,
		       s.created_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN lecturers l ON s.advisor_id = l.id
		LEFT JOIN users lu ON l.user_id = lu.id
		WHERE u.is_active = true
		ORDER BY u.full_name
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []models.StudentResponse
	for rows.Next() {
		var s models.StudentResponse
		var advisorName, lecturerID sql.NullString
		
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.FullName,
			&s.StudentID, &s.ProgramStudy, &s.AcademicYear,
			&s.AdvisorID, &lecturerID, &advisorName,
			&s.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		
		if advisorName.Valid {
			s.AdvisorName = advisorName.String
		}
		
		students = append(students, s)
	}

	// Get total count
	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE u.is_active = true
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepo) GetAllByAdvisorID(advisorID uuid.UUID, page, limit int) ([]models.Student, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, 
		       advisor_id, created_at
		FROM students 
		WHERE advisor_id=$1 
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, advisorID, limit, offset)
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
	`, advisorID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepo) GetAdvisorless(page, limit int) ([]models.Student, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, 
		       advisor_id, created_at
		FROM students 
		WHERE advisor_id IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
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
		WHERE advisor_id IS NULL
	`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepo) SearchByName(name string, page, limit int) ([]models.StudentResponse, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	searchPattern := "%" + name + "%"
	rows, err := r.DB.Query(`
		SELECT s.id, s.user_id, u.full_name, u.email, u.username,
		       s.student_id, s.program_study, s.academic_year, 
		       s.advisor_id, s.created_at
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE u.is_active = true AND u.full_name ILIKE $1
		ORDER BY u.full_name
		LIMIT $2 OFFSET $3
	`, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []models.StudentResponse
	for rows.Next() {
		var s models.StudentResponse
		if err := rows.Scan(
			&s.ID, &s.UserID, &s.FullName, &s.Email, &s.Username,
			&s.StudentID, &s.ProgramStudy, &s.AcademicYear,
			&s.AdvisorID, &s.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		students = append(students, s)
	}

	// Get total count
	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM students s
		JOIN users u ON s.user_id = u.id
		WHERE u.is_active = true AND u.full_name ILIKE $1
	`, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepo) GetByProgramStudy(programStudy string, page, limit int) ([]models.Student, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, 
		       advisor_id, created_at
		FROM students 
		WHERE program_study=$1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, programStudy, limit, offset)
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

	// Get total count
	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM students
		WHERE program_study=$1
	`, programStudy).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepo) GetByAcademicYear(academicYear string, page, limit int) ([]models.Student, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, user_id, student_id, program_study, academic_year, 
		       advisor_id, created_at
		FROM students 
		WHERE academic_year=$1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, academicYear, limit, offset)
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

	// Get total count
	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM students
		WHERE academic_year=$1
	`, academicYear).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}

func (r *studentRepo) GetStudentsCountByAdvisor(advisorID uuid.UUID) (int, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*) 
		FROM students 
		WHERE advisor_id=$1
	`, advisorID).Scan(&count)
	
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *studentRepo) GetStudentsCountByProgramStudy() (map[string]int, error) {
	rows, err := r.DB.Query(`
		SELECT program_study, COUNT(*) as count
		FROM students
		GROUP BY program_study
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var programStudy string
		var count int
		if err := rows.Scan(&programStudy, &count); err != nil {
			return nil, err
		}
		result[programStudy] = count
	}

	return result, nil
}