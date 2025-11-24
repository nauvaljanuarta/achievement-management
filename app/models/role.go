package models

import (
	"time"
	"github.com/google/uuid"
)


type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"` // Admin, Mahasiswa, Dosen Wali
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

type Permission struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"` // contoh: achievement:create
	Resource    string    `json:"resource" db:"resource"`
	Action      string    `json:"action" db:"action"`
	Description string    `json:"description" db:"description"`
}

type RolePermission struct {
	RoleID       uuid.UUID `json:"roleId" db:"role_id"`
	PermissionID uuid.UUID `json:"permissionId" db:"permission_id"`
}