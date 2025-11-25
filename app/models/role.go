package models

import (
	"time"
	"github.com/google/uuid"
)


type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"` // Admin, Mahasiswa, dosen Wali
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

