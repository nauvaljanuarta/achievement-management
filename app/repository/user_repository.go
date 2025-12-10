package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"achievement-backend/app/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	// Basic CRUD
	GetByID(id uuid.UUID) (*models.User, error)           // Hanya active users
	GetByEmail(email string) (*models.User, error)        // Hanya active users
	GetByUsername(username string) (*models.User, error)  // Hanya active users
	GetByUsernameOrEmail(identifier string) (*models.User, error) // Hanya active users
	Create(user *models.User) (uuid.UUID, error)
	Update(id uuid.UUID, req *models.UpdateUserRequest) error
	UpdatePassword(id uuid.UUID, hashedPassword string) error
	SoftDelete(id uuid.UUID) error
	HardDelete(id uuid.UUID) error
	
	GetAll(page, limit int) ([]models.User, int, error) 
	
	GetInactiveUsers(page, limit int) ([]models.User, int, error) 
	
	GetAllWithInactive(page, limit int) ([]models.User, int, error) 
	
	SearchByName(name string, page, limit int) ([]models.User, int, error) 
	GetByRole(roleID uuid.UUID, page, limit int) ([]models.User, int, error) 
	
	GetUsersCountByRole() (map[uuid.UUID]int, error) 
	GetTotalActiveCount() (int, error) 
	GetTotalInactiveCount() (int, error) 
}

type userRepo struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{DB: db}
}

func (r *userRepo) GetByID(id uuid.UUID) (*models.User, error) {
	var u models.User
	err := r.DB.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role_id, 
		       is_active, created_at, updated_at
		FROM users
		WHERE id=$1 AND is_active=true
	`, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.RoleID,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) GetByEmail(email string) (*models.User, error) {
	var u models.User
	err := r.DB.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role_id, 
		       is_active, created_at, updated_at
		FROM users
		WHERE email=$1 AND is_active=true
	`, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.RoleID,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) GetByUsername(username string) (*models.User, error) {
	var u models.User
	err := r.DB.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role_id, 
		       is_active, created_at, updated_at
		FROM users
		WHERE username=$1 AND is_active=true
	`, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.RoleID,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) GetByUsernameOrEmail(identifier string) (*models.User, error) {
	var u models.User
	err := r.DB.QueryRow(`
		SELECT id, username, email, password_hash, full_name, role_id, 
		       is_active, created_at, updated_at
		FROM users
		WHERE (username=$1 OR email=$1) AND is_active=true
	`, identifier).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.RoleID,
		&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *userRepo) Create(user *models.User) (uuid.UUID, error) {
	_, err := r.DB.Exec(`
		INSERT INTO users (id, username, email, password_hash, full_name, role_id, 
		                  is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`, 
		user.ID,          
		user.Username,     
		user.Email,          
		user.PasswordHash, 
		user.FullName,     
		user.RoleID,       
		user.IsActive,     
	)

	if err != nil {
		return uuid.Nil, err
	}

	return user.ID, nil
}

func (r *userRepo) Update(id uuid.UUID, req *models.UpdateUserRequest) error {
	query := `UPDATE users SET updated_at=$1`
	params := []interface{}{time.Now()}
	paramIndex := 2
	
	if req.FullName != nil {
		query += `, full_name=$` + fmt.Sprintf("%d", paramIndex)
		params = append(params, *req.FullName)
		paramIndex++
	}
	if req.Email != nil {
		query += `, email=$` + fmt.Sprintf("%d", paramIndex)
		params = append(params, *req.Email)
		paramIndex++
	}
	if req.IsActive != nil {
		query += `, is_active=$` + fmt.Sprintf("%d", paramIndex)
		params = append(params, *req.IsActive)
		paramIndex++
	}
	if req.RoleID != nil {
		query += `, role_id=$` + fmt.Sprintf("%d", paramIndex)
		params = append(params, *req.RoleID)
		paramIndex++
	}
	
	query += ` WHERE id=$` + fmt.Sprintf("%d", paramIndex)
	params = append(params, id)
	
	_, err := r.DB.Exec(query, params...)
	return err
}

func (r *userRepo) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	_, err := r.DB.Exec(`
		UPDATE users 
		SET password_hash=$1, updated_at=$2
		WHERE id=$3
	`, hashedPassword, time.Now(), id)
	return err
}

func (r *userRepo) SoftDelete(id uuid.UUID) error {
	_, err := r.DB.Exec(`
		UPDATE users 
		SET is_active=false, updated_at=$1 
		WHERE id=$2
	`, time.Now(), id)
	return err
}

func (r *userRepo) GetAll(page, limit int) ([]models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, username, email, full_name, role_id, 
		       is_active, created_at, updated_at
		FROM users
		WHERE is_active=true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleID,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	total, err := r.GetTotalActiveCount()
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepo) GetInactiveUsers(page, limit int) ([]models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, username, email, full_name, role_id, 
		       is_active, created_at, updated_at
		FROM users
		WHERE is_active=false
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleID,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	total, err := r.GetTotalInactiveCount()
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepo) GetAllWithInactive(page, limit int) ([]models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, username, email, full_name, role_id, 
		       is_active, created_at, updated_at
		FROM users
		ORDER BY is_active DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleID,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	// Get total ALL count
	var total int
	err = r.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepo) SearchByName(name string, page, limit int) ([]models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	searchPattern := "%" + name + "%"
	rows, err := r.DB.Query(`
		SELECT id, username, email, full_name, role_id, 
		       is_active, created_at, updated_at
		FROM users
		WHERE is_active=true AND full_name ILIKE $1
		ORDER BY full_name
		LIMIT $2 OFFSET $3
	`, searchPattern, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleID,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	// Get total count for search
	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM users
		WHERE is_active=true AND full_name ILIKE $1
	`, searchPattern).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepo) GetByRole(roleID uuid.UUID, page, limit int) ([]models.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, username, email, full_name, role_id, 
		       is_active, created_at, updated_at
		FROM users
		WHERE role_id=$1 AND is_active=true
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, roleID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleID,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	var total int
	err = r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM users
		WHERE role_id=$1 AND is_active=true
	`, roleID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepo) GetUsersCountByRole() (map[uuid.UUID]int, error) {
	rows, err := r.DB.Query(`
		SELECT role_id, COUNT(*) as count
		FROM users
		WHERE is_active=true
		GROUP BY role_id
		ORDER BY count DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]int)
	for rows.Next() {
		var roleID uuid.UUID
		var count int
		if err := rows.Scan(&roleID, &count); err != nil {
			return nil, err
		}
		result[roleID] = count
	}

	return result, nil
}

func (r *userRepo) GetTotalActiveCount() (int, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE is_active=true`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepo) GetTotalInactiveCount() (int, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE is_active=false`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepo) HardDelete(id uuid.UUID) error {
	_, err := r.DB.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}