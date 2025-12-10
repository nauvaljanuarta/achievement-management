package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"achievement-backend/app/models"
	"github.com/google/uuid"
)

type RoleRepository interface {
	GetByID(id uuid.UUID) (*models.Role, error)
	GetByName(name string) (*models.Role, error)
	GetAll(page, limit int) ([]models.Role, int, error)
	GetTotalCount() (int, error)
	GetPermissionsByRoleID(roleID uuid.UUID) ([]models.Permission, error)
	GetPermissionNamesByRoleID(roleID uuid.UUID) ([]string, error)
	AssignPermission(roleID, permissionID uuid.UUID) error
	RemovePermission(roleID, permissionID uuid.UUID) error
}

type roleRepo struct {
	DB *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepo{DB: db}
}

func (r *roleRepo) GetByID(id uuid.UUID) (*models.Role, error) {
	var role models.Role
	err := r.DB.QueryRow(`
		SELECT id, name, description, created_at
		FROM roles
		WHERE id=$1
	`, id).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
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
	err := r.DB.QueryRow(`
		SELECT id, name, description, created_at
		FROM roles
		WHERE name=$1
	`, name).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *roleRepo) GetAll(page, limit int) ([]models.Role, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	rows, err := r.DB.Query(`
		SELECT id, name, description, created_at
		FROM roles
		ORDER BY name
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, 0, err
		}
		roles = append(roles, role)
	}

	total, err := r.GetTotalCount()
	if err != nil {
		return nil, 0, err
	}

	return roles, total, nil
}

func (r *roleRepo) GetTotalCount() (int, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM roles`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *roleRepo) GetPermissionsByRoleID(roleID uuid.UUID) ([]models.Permission, error) {
	rows, err := r.DB.Query(`
		SELECT p.id, p.name, p.resource, p.action, p.description
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.resource, p.action
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []models.Permission
	for rows.Next() {
		var p models.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &p.Description); err != nil {
			return nil, err
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}

func (r *roleRepo) GetPermissionNamesByRoleID(roleID uuid.UUID) ([]string, error) {
	rows, err := r.DB.Query(`
		SELECT p.name
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1
		ORDER BY p.name
	`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permName string
		if err := rows.Scan(&permName); err != nil {
			return nil, err
		}
		permissions = append(permissions, permName)
	}
	return permissions, nil
}

func (r *roleRepo) AssignPermission(roleID, permissionID uuid.UUID) error {
	var exists bool
	err := r.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM role_permissions 
			WHERE role_id=$1 AND permission_id=$2
		)
	`, roleID, permissionID).Scan(&exists)
	
	if err != nil {
		return fmt.Errorf("error checking existing permission: %w", err)
	}
	
	if exists {
		return fmt.Errorf("permission already assigned to role")
	}
	
	var roleExists, permExists bool
	err = r.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM roles WHERE id=$1)`, roleID).Scan(&roleExists)
	if err != nil {
		return fmt.Errorf("error checking role: %w", err)
	}
	
	if !roleExists {
		return fmt.Errorf("role not found")
	}
	
	err = r.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM permissions WHERE id=$1)`, permissionID).Scan(&permExists)
	if err != nil {
		return fmt.Errorf("error checking permission: %w", err)
	}
	
	if !permExists {
		return fmt.Errorf("permission not found")
	}
	
	_, err = r.DB.Exec(`
		INSERT INTO role_permissions (role_id, permission_id)
		VALUES ($1, $2)
	`, roleID, permissionID)
	
	if err != nil {
		return fmt.Errorf("error assigning permission: %w", err)
	}
	
	return nil
}

func (r *roleRepo) RemovePermission(roleID, permissionID uuid.UUID) error {
	result, err := r.DB.Exec(`
		DELETE FROM role_permissions 
		WHERE role_id=$1 AND permission_id=$2
	`, roleID, permissionID)
	
	if err != nil {
		return fmt.Errorf("error removing permission: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("permission not found for this role")
	}
	
	return nil
}

