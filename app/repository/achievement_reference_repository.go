package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"achievement-backend/app/models"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type AchievementReferenceRepository interface {
	// Basic CRUD
	Create(ctx context.Context, ref *models.AchievementReference) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error)
	FindByStudentID(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error)
	FindByMongoID(ctx context.Context, mongoID string) (*models.AchievementReference, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedBy *uuid.UUID, rejectionNote *string) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	// For Dosen Wali
	FindByAdvisorID(ctx context.Context, advisorID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error)
	// For Admin
	FindAll(ctx context.Context, status string, page, limit int) ([]*models.AchievementReference, int, error)
	// Status transitions
	SubmitForVerification(ctx context.Context, id uuid.UUID) error
	VerifyAchievement(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error
	RejectAchievement(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID, rejectionNote string) error
	// Statistics
	CountByStatus(ctx context.Context, studentID uuid.UUID) (map[string]int, error)
	CountByStudentAndStatus(ctx context.Context, studentID uuid.UUID, status string) (int, error)
	// Get student IDs for advisor (helper)
	GetStudentIDsByAdvisor(ctx context.Context, advisorID uuid.UUID) ([]uuid.UUID, error)
}

type achievementReferenceRepo struct {
	DB *sql.DB
}

func NewAchievementReferenceRepository(db *sql.DB) AchievementReferenceRepository {
	return &achievementReferenceRepo{DB: db}
}

func (r *achievementReferenceRepo) Create(ctx context.Context, ref *models.AchievementReference) error {
	query := `
		INSERT INTO achievement_references 
		(id, student_id, mongo_achievement_id, status, submitted_at, 
		 verified_at, verified_by, rejection_note, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	
	_, err := r.DB.ExecContext(ctx, query,
		ref.ID,
		ref.StudentID,
		ref.MongoAchievementID,
		ref.Status,
		ref.SubmittedAt,
		ref.VerifiedAt,
		ref.VerifiedBy,
		ref.RejectionNote,
		ref.CreatedAt,
		ref.UpdatedAt,
	)
	
	return err
}

func (r *achievementReferenceRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		WHERE id = $1
	`
	
	var ref models.AchievementReference
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &ref, nil
}

func (r *achievementReferenceRepo) FindByStudentID(ctx context.Context, studentID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	
	// Build query with optional status filter
	baseQuery := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		WHERE student_id = $1
	`
	countQuery := `SELECT COUNT(*) FROM achievement_references WHERE student_id = $1`
	
	params := []interface{}{studentID}
	paramCount := 2
	
	if status == "deleted" {
		// Khusus minta deleted
		baseQuery += fmt.Sprintf(` AND status = $%d`, paramCount)
		countQuery += fmt.Sprintf(` AND status = $%d`, paramCount)
		params = append(params, status)
		paramCount++
	} else if status != "" {
		// Filter status tertentu, exclude deleted
		baseQuery += fmt.Sprintf(` AND status = $%d AND status != 'deleted'`, paramCount)
		countQuery += fmt.Sprintf(` AND status = $%d AND status != 'deleted'`, paramCount)
		params = append(params, status)
		paramCount++
	} else {
		// Default: exclude deleted
		baseQuery += ` AND status != 'deleted'`
		countQuery += ` AND status != 'deleted'`
	}
	
	baseQuery += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	params = append(params, limit, offset)
	
	// Get total count
	var total int
	err := r.DB.QueryRowContext(ctx, countQuery, params[:paramCount-1]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	rows, err := r.DB.QueryContext(ctx, baseQuery, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var references []*models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.VerifiedAt,
			&ref.VerifiedBy,
			&ref.RejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		references = append(references, &ref)
	}
	
	return references, total, nil
}

func (r *achievementReferenceRepo) FindByMongoID(ctx context.Context, mongoID string) (*models.AchievementReference, error) {
	query := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
		WHERE mongo_achievement_id = $1
	`
	
	var ref models.AchievementReference
	err := r.DB.QueryRowContext(ctx, query, mongoID).Scan(
		&ref.ID,
		&ref.StudentID,
		&ref.MongoAchievementID,
		&ref.Status,
		&ref.SubmittedAt,
		&ref.VerifiedAt,
		&ref.VerifiedBy,
		&ref.RejectionNote,
		&ref.CreatedAt,
		&ref.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return &ref, nil
}

func (r *achievementReferenceRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, verifiedBy *uuid.UUID, rejectionNote *string) error {
	query := `
		UPDATE achievement_references 
		SET status = $1, 
		    verified_by = $2,
		    rejection_note = $3,
		    updated_at = $4
	`
	params := []interface{}{status, verifiedBy, rejectionNote, time.Now()}
	paramCount := 5
	
	switch status {
	case "submitted":
		query += `, submitted_at = $` + fmt.Sprintf("%d", paramCount)
		params = append(params, time.Now())
		paramCount++
	case "verified":
		query += `, verified_at = $` + fmt.Sprintf("%d", paramCount)
		params = append(params, time.Now())
		paramCount++
	}
	
	query += ` WHERE id = $` + fmt.Sprintf("%d", paramCount)
	params = append(params, id)
	
	_, err := r.DB.ExecContext(ctx, query, params...)
	return err
}

func (r *achievementReferenceRepo) Update(ctx context.Context, ref *models.AchievementReference) error {
	query := `
		UPDATE achievement_references 
		SET status = $1,
		    submitted_at = $2,
		    verified_at = $3,
		    verified_by = $4,
		    rejection_note = $5,
		    updated_at = $6
		WHERE id = $7
	`
	
	_, err := r.DB.ExecContext(ctx, query,
		ref.Status,
		ref.SubmittedAt,
		ref.VerifiedAt,
		ref.VerifiedBy,
		ref.RejectionNote,
		ref.UpdatedAt,
		ref.ID,
	)
	return err
}

func (r *achievementReferenceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM achievement_references WHERE id = $1`
	_, err := r.DB.ExecContext(ctx, query, id)
	return err
}

func (r *achievementReferenceRepo) SoftDelete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE achievement_references 
		SET status = 'deleted', 
		    updated_at = $1
		WHERE id = $2 AND status = 'draft'  -- Hanya draft yang bisa di-delete
	`
	result, err := r.DB.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("achievement not found or not in draft status")
	}
	
	return nil
}

func (r *achievementReferenceRepo) FindByAdvisorID(ctx context.Context, advisorID uuid.UUID, status string, page, limit int) ([]*models.AchievementReference, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	
	// First, get student IDs for this advisor
	studentIDs, err := r.GetStudentIDsByAdvisor(ctx, advisorID)
	if err != nil {
		return nil, 0, err
	}
	
	if len(studentIDs) == 0 {
		return []*models.AchievementReference{}, 0, nil
	}
	
	studentIDsArray := pq.Array(studentIDs)
	
	baseQuery := `
		SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, 
		       ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note,
		       ar.created_at, ar.updated_at
		FROM achievement_references ar
		WHERE ar.student_id = ANY($1)
	`
	countQuery := `SELECT COUNT(*) FROM achievement_references WHERE student_id = ANY($1)`
	
	params := []interface{}{studentIDsArray}
	paramCount := 2
	
	if status == "deleted" {
		// Khusus minta deleted
		baseQuery += fmt.Sprintf(` AND ar.status = $%d`, paramCount)
		countQuery += fmt.Sprintf(` AND status = $%d`, paramCount)
		params = append(params, status)
		paramCount++
	} else if status != "" {
		// Filter status tertentu, exclude deleted
		baseQuery += fmt.Sprintf(` AND ar.status = $%d AND ar.status != 'deleted'`, paramCount)
		countQuery += fmt.Sprintf(` AND status = $%d AND status != 'deleted'`, paramCount)
		params = append(params, status)
		paramCount++
	} else {
		// Default: exclude deleted
		baseQuery += ` AND ar.status != 'deleted'`
		countQuery += ` AND status != 'deleted'`
	}
	
	baseQuery += ` ORDER BY ar.created_at DESC LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	params = append(params, limit, offset)
	
	// Get total count
	var total int
	err = r.DB.QueryRowContext(ctx, countQuery, params[:len(params)-2]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	rows, err := r.DB.QueryContext(ctx, baseQuery, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var references []*models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.VerifiedAt,
			&ref.VerifiedBy,
			&ref.RejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		references = append(references, &ref)
	}
	
	return references, total, nil
}

func (r *achievementReferenceRepo) FindAll(ctx context.Context, status string, page, limit int) ([]*models.AchievementReference, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	
	// Build query
	baseQuery := `
		SELECT id, student_id, mongo_achievement_id, status, 
		       submitted_at, verified_at, verified_by, rejection_note,
		       created_at, updated_at
		FROM achievement_references
	`
	countQuery := `SELECT COUNT(*) FROM achievement_references`
	
	params := []interface{}{}
	paramCount := 1
	
	if status == "deleted" {
		baseQuery += ` WHERE status = 'deleted'`
		countQuery += ` WHERE status = 'deleted'`
	} else if status != "" {
		// Filter status tertentu, exclude deleted
		baseQuery += ` WHERE status = $1 AND status != 'deleted'`
		countQuery += ` WHERE status = $1 AND status != 'deleted'`
		params = append(params, status)
		paramCount++
	} else {
		// Default: exclude deleted
		baseQuery += ` WHERE status != 'deleted'`
		countQuery += ` WHERE status != 'deleted'`
	}
	
	baseQuery += ` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", paramCount) + ` OFFSET $` + fmt.Sprintf("%d", paramCount+1)
	params = append(params, limit, offset)
	
	// Get total count
	var total int
	err := r.DB.QueryRowContext(ctx, countQuery, params[:paramCount-1]...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	rows, err := r.DB.QueryContext(ctx, baseQuery, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	var references []*models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.SubmittedAt,
			&ref.VerifiedAt,
			&ref.VerifiedBy,
			&ref.RejectionNote,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		references = append(references, &ref)
	}
	
	return references, total, nil
}

func (r *achievementReferenceRepo) SubmitForVerification(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE achievement_references 
		SET status = 'submitted', 
		    submitted_at = $1,
		    updated_at = $2
		WHERE id = $3 AND status = 'draft'
	`
	
	result, err := r.DB.ExecContext(ctx, query, time.Now(), time.Now(), id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("achievement not found or not in draft status")
	}
	
	return nil
}

func (r *achievementReferenceRepo) VerifyAchievement(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) error {
	query := `
		UPDATE achievement_references 
		SET status = 'verified', 
		    verified_at = $1,
		    verified_by = $2,
		    updated_at = $3
		WHERE id = $4 AND status = 'submitted'
	`
	
	result, err := r.DB.ExecContext(ctx, query, time.Now(), verifiedBy, time.Now(), id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("achievement not found or not in submitted status")
	}
	
	return nil
}

func (r *achievementReferenceRepo) RejectAchievement(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID, rejectionNote string) error {
	query := `
		UPDATE achievement_references 
		SET status = 'rejected', 
		    verified_by = $1,
		    rejection_note = $2,
		    updated_at = $3
		WHERE id = $4 AND status = 'submitted'
	`
	
	result, err := r.DB.ExecContext(ctx, query, verifiedBy, rejectionNote, time.Now(), id)
	if err != nil {
		return err
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("achievement not found or not in submitted status")
	}
	
	return nil
}

func (r *achievementReferenceRepo) CountByStatus(ctx context.Context, studentID uuid.UUID) (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) 
		FROM achievement_references 
		WHERE student_id = $1
		GROUP BY status
	`
	
	rows, err := r.DB.QueryContext(ctx, query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	result := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		result[status] = count
	}
	
	// Ensure all statuses are in the map
	statuses := []string{"draft", "submitted", "verified", "rejected", "deleted"}
	for _, status := range statuses {
		if _, exists := result[status]; !exists {
			result[status] = 0
		}
	}
	
	return result, nil
}

func (r *achievementReferenceRepo) CountByStudentAndStatus(ctx context.Context, studentID uuid.UUID, status string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM achievement_references 
		WHERE student_id = $1 AND status = $2
	`
	
	var count int
	err := r.DB.QueryRowContext(ctx, query, studentID, status).Scan(&count)
	return count, err
}

func (r *achievementReferenceRepo) GetStudentIDsByAdvisor(ctx context.Context, advisorID uuid.UUID) ([]uuid.UUID, error) {
	query := `
		SELECT id 
		FROM students 
		WHERE advisor_id = $1
	`
	
	rows, err := r.DB.QueryContext(ctx, query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var studentIDs []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, id)
	}
	
	return studentIDs, nil
}