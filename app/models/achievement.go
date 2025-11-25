package models

import (
	"time"
	"github.com/google/uuid"
)


type AchievementReference struct {
	ID                 uuid.UUID  `json:"id"`
	StudentID          uuid.UUID  `json:"student_id"`
	MongoAchievementID string     `json:"mongo_achievement_id"`
	Status             string     `json:"status"`
	SubmittedAt        *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	VerifiedBy         *uuid.UUID `json:"verified_by,omitempty"`
	RejectionNote      string     `json:"rejection_note"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
