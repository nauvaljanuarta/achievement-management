package models

import "github.com/google/uuid"


type AchievementStatistics struct {
	TotalAchievements int            `json:"total_achievements"`
	ByType           map[string]int `json:"by_type"`
	ByPeriod         map[string]int `json:"by_period"`
	TopTags          []string       `json:"top_tags"`
	AveragePoints    float64        `json:"average_points"`
}

type AchievementSummary struct {
	StudentID            uuid.UUID            `json:"student_id"`
	TotalAchievements    int                  `json:"total_achievements"`
	TotalPoints          int                  `json:"total_points"`
	AchievementByType    map[string]int       `json:"achievement_by_type"`
	AchievementByStatus  map[string]int       `json:"achievement_by_status"`
}

type StudentAchievementSummary struct {
	StudentID   uuid.UUID `json:"student_id"`
	FullName    string    `json:"full_name"`
	TotalPoints int       `json:"total_points"`
	VerifiedCount int     `json:"verified_count"`
}