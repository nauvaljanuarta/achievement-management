package repository

import (
	"database/sql"
	"errors"
	"achievement-backend/app/models"
)


type RoleRepository interface {
	GetByID(id string) (*models.Role, error)
	GetByName(name string) (*models.Role, error)
	GetAll() ([]models.Role, error)
}

type roleRepo struct {
	DB *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepo{DB: db}
}

func (r *roleRepo) GetByID(id string) (*models.Role, error) {
	var role models.Role
	err := r.DB.QueryRow(`SELECT id, name, description, created_at FROM roles WHERE id=$1`, id).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) GetByName(name string) (*models.Role, error) {
	var role models.Role
	err := r.DB.QueryRow(`SELECT id, name, description, created_at FROM roles WHERE name=$1`, name).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) GetAll() ([]models.Role, error) {
	rows, err := r.DB.Query(`SELECT id, name, description, created_at FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var r models.Role
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, r)
	}
	return roles, nil
}
