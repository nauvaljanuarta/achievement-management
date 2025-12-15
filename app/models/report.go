package models

import "github.com/google/uuid"

type AchievementStats struct {
	ByType             map[string]int          `json:"by_type"`
	ByPeriod           map[string]int          `json:"by_period"`
	ByCompetitionLevel map[string]int          `json:"by_competition_level"`
	TopStudents        []StudentAchievementSum `json:"top_students"`
}

type StudentAchievementSum struct {
	StudentID   uuid.UUID `json:"student_id"`
	StudentName string    `json:"student_name"`
	TotalCount  int       `json:"total_count"`
	TotalPoints int       `json:"total_points"`
}