// app/repository/report_repository.go
package repository

import (
	"context"
	"time"

	"achievement-backend/app/models"
	"achievement-backend/database"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"database/sql"
)

type ReportRepository interface {
	GetAchievementStats(ctx context.Context, actorID uuid.UUID, roleName string, startDate, endDate *time.Time) (*models.AchievementStats, error)
}

type reportRepo struct{}

func NewReportRepository() ReportRepository {
	return &reportRepo{}
}

// app/repository/report_repository.go (FIXED)
func (r *reportRepo) GetAchievementStats(ctx context.Context, actorID uuid.UUID, roleName string, startDate, endDate *time.Time) (*models.AchievementStats, error) {
	stats := &models.AchievementStats{
		ByType:             make(map[string]int),
		ByPeriod:           make(map[string]int),
		ByCompetitionLevel: make(map[string]int),
		TopStudents:        []models.StudentAchievementSum{},
	}

	var whereClause string
	var queryParams []interface{}

	switch roleName {
	case "Admin":
		whereClause = "WHERE ar.status != 'deleted'"
	case "Dosen Wali":
		whereClause = `
			WHERE ar.status != 'deleted' 
			AND ar.student_id IN (
				SELECT id FROM students WHERE advisor_id = $1
			)
		`
		queryParams = append(queryParams, actorID)
	case "Mahasiswa":
		whereClause = "WHERE ar.status != 'deleted' AND ar.student_id = $1"
		queryParams = append(queryParams, actorID)
	}

	// Add date filter
	if startDate != nil && endDate != nil {
		if len(queryParams) > 0 {
			whereClause += " AND ar.created_at BETWEEN $2 AND $3"
			queryParams = append(queryParams, startDate, endDate)
		} else {
			whereClause += " WHERE ar.created_at BETWEEN $1 AND $2"
			queryParams = append(queryParams, startDate, endDate)
		}
	}

	// 1. Total prestasi per periode
	periodQuery := `
		SELECT 
			TO_CHAR(ar.created_at, 'YYYY-MM') as period,
			COUNT(*) as count
		FROM achievement_references ar
		` + whereClause + `
		GROUP BY TO_CHAR(ar.created_at, 'YYYY-MM')
		ORDER BY period
	`

	rows, _ := database.PgDB.QueryContext(ctx, periodQuery, queryParams...)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var period string
			var count int
			if err := rows.Scan(&period, &count); err == nil {
				stats.ByPeriod[period] = count
			}
		}
	}

	// 2. Get student IDs for MongoDB query
	studentIDsQuery := `SELECT DISTINCT ar.student_id FROM achievement_references ar ` + whereClause
	studentRows, err := database.PgDB.QueryContext(ctx, studentIDsQuery, queryParams...)
	if err != nil {
		return stats, err
	}
	defer studentRows.Close()

	var studentIDStrings []string
	for studentRows.Next() {
		var studentID uuid.UUID
		if err := studentRows.Scan(&studentID); err == nil {
			studentIDStrings = append(studentIDStrings, studentID.String())
		}
	}

	if len(studentIDStrings) == 0 {
		return stats, nil
	}

	// 3. Total prestasi per tipe (dari MongoDB)
	typePipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDStrings}}},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, _ := database.MongoDB.Collection("achievements").Aggregate(ctx, typePipeline)
	if cursor != nil {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var result struct {
				Type  string `bson:"_id"`
				Count int    `bson:"count"`
			}
			if err := cursor.Decode(&result); err == nil {
				stats.ByType[result.Type] = result.Count
			}
		}
	}

	// 4. Distribusi tingkat kompetisi
	compPipeline := bson.A{
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "studentId", Value: bson.D{{Key: "$in", Value: studentIDStrings}}},
			{Key: "achievementType", Value: "competition"},
		}}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$details.competitionLevel"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	compCursor, _ := database.MongoDB.Collection("achievements").Aggregate(ctx, compPipeline)
	if compCursor != nil {
		defer compCursor.Close(ctx)
		for compCursor.Next(ctx) {
			var result struct {
				Level string `bson:"_id"`
				Count int    `bson:"count"`
			}
			if err := compCursor.Decode(&result); err == nil {
				stats.ByCompetitionLevel[result.Level] = result.Count
			}
		}
	}

	// 5. Top mahasiswa berprestasi (Admin & Dosen Wali only) - FIXED!
	if roleName == "Admin" || roleName == "Dosen Wali" {
		// Query sederhana: hitung prestasi verified per student
		topStudentsQuery := `
			SELECT 
				s.id, 
				u.full_name,
				COUNT(ar.id) as achievement_count,
				COUNT(ar.id) as total_count -- Untuk sementara, poin dari MongoDB nanti
			FROM students s
			JOIN users u ON s.user_id = u.id
			LEFT JOIN achievement_references ar ON s.id = ar.student_id 
			WHERE ar.status = 'verified'
		`

		// Add filter untuk Dosen Wali
		if roleName == "Dosen Wali" {
			topStudentsQuery += ` AND s.advisor_id = $1`
		}

		// Add GROUP BY dan ORDER
		topStudentsQuery += `
			GROUP BY s.id, u.full_name
			ORDER BY achievement_count DESC
			LIMIT 10
		`

		var rows *sql.Rows
		if roleName == "Dosen Wali" {
			rows, _ = database.PgDB.QueryContext(ctx, topStudentsQuery, actorID)
		} else {
			rows, _ = database.PgDB.QueryContext(ctx, topStudentsQuery)
		}

		if rows != nil {
			defer rows.Close()
			for rows.Next() {
				var student models.StudentAchievementSum
				var fullName string
				var achievementCount int
				if err := rows.Scan(&student.StudentID, &fullName, &achievementCount, &student.TotalCount); err == nil {
					student.StudentName = fullName
					student.TotalCount = achievementCount
					student.TotalPoints = achievementCount * 10 // Default points sementara
					stats.TopStudents = append(stats.TopStudents, student)
				}
			}
		}
		
	}

	return stats, nil
}