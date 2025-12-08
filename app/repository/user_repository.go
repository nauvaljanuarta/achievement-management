package repository

import (
	"database/sql"
	"errors"

	"achievement-backend/app/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Create(user models.User) (uuid.UUID, error)
	Update(user models.User) error
	Delete(id uuid.UUID) error
	GetAll() ([]models.User, error)
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
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE id=$1
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
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE email=$1
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

func (r *userRepo) Create(user models.User) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.DB.QueryRow(`
		INSERT INTO users (username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id
	`, user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive).Scan(&id)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *userRepo) Update(user models.User) error {
	_, err := r.DB.Exec(`
		UPDATE users
		SET username=$1, email=$2, password_hash=$3, full_name=$4, role_id=$5, is_active=$6, updated_at=NOW()
		WHERE id=$7
	`, user.Username, user.Email, user.PasswordHash, user.FullName, user.RoleID, user.IsActive, user.ID)
	return err
}

func (r *userRepo) Delete(id uuid.UUID) error {
	_, err := r.DB.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}

func (r *userRepo) GetAll() ([]models.User, error) {
	rows, err := r.DB.Query(`
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.FullName, &u.RoleID,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
