package models

import (
	"time"
	"github.com/google/uuid"
)


type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"` 
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

type RolePermission struct {
	RoleID       uuid.UUID `json:"roleId" db:"role_id"`           
	PermissionID uuid.UUID `json:"permissionId" db:"permission_id"` 
}
