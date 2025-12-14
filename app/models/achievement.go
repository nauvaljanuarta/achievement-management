package models

import (
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Achievement struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StudentID       string        `bson:"studentId" json:"student_id"`
	AchievementType string             `bson:"achievementType" json:"achievement_type"`
	Title           string             `bson:"title" json:"title"`
	Description     string             `bson:"description" json:"description"`

	Details AchievementDetails `bson:"details" json:"details"`

	Attachments []Attachment `bson:"attachments" json:"attachments"`
	Tags        []string     `bson:"tags" json:"tags"`
	Points      int          `bson:"points" json:"points"`

	CreatedAt time.Time `bson:"createdAt" json:"created_at"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updated_at"`
}

type CreateAchievementRequest struct {
	AchievementType string             `json:"achievement_type" validate:"required"`
	Title           string             `json:"title" validate:"required"`
	Description     string             `json:"description"`
	Details         AchievementDetails `json:"details"`
	Attachments     []Attachment       `json:"attachments"`
	Tags            []string           `json:"tags"`
	Points          int                `json:"points"`
	StudentID       *uuid.UUID         `json:"student_id,omitempty"`
}
