// app/service/report_service.go
package service

import (
	"context"
	"time"

	"achievement-backend/app/models"
	"achievement-backend/app/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ReportService struct {
	reportRepo  repository.ReportRepository
	userRepo    repository.UserRepository
	studentRepo repository.StudentRepository
	lecturerRepo repository.LecturerRepository
	roleRepo    repository.RoleRepository
}

func NewReportService(
	reportRepo repository.ReportRepository,
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	roleRepo repository.RoleRepository,
) *ReportService {
	return &ReportService{
		reportRepo:  reportRepo,
		userRepo:    userRepo,
		studentRepo: studentRepo,
		lecturerRepo: lecturerRepo,
		roleRepo:    roleRepo,
	}
}

// GetStatistics godoc
// @Summary Get achievement statistics
// @Description
// Mengambil statistik prestasi berdasarkan role pengguna:
// - Admin: statistik keseluruhan
// - Dosen Wali: statistik mahasiswa bimbingannya
// - Mahasiswa: statistik prestasinya sendiri
//
// Mendukung filter rentang tanggal menggunakan start_date dan end_date.
//
// @Tags Report
// @Security BearerAuth
// @Produce json
//
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
//
// @Success 200 {object} map[string]interface{} "Statistics data"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 500 {object} map[string]string "Failed to generate statistics"
//
// @Router /reports/statistics [get]
func (s *ReportService) GetStatistics(c *fiber.Ctx) error {
	currentUser, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	role, err := s.roleRepo.GetByID(currentUser.RoleID)
	if err != nil || role == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get user role"})
	}

	// Parse dates
	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &parsed
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &parsed
		}
	}

	// Determine actor ID
	var actorID uuid.UUID
	switch role.Name {
	case "Admin":
		actorID = currentUser.ID
	case "Dosen Wali":
		lecturer, err := s.lecturerRepo.GetByUserID(currentUser.ID)
		if err != nil || lecturer == nil {
			return c.Status(403).JSON(fiber.Map{"error": "You are not registered as a lecturer"})
		}
		actorID = lecturer.ID
	case "Mahasiswa":
		student, err := s.studentRepo.GetByUserID(currentUser.ID)
		if err != nil || student == nil {
			return c.Status(403).JSON(fiber.Map{"error": "You are not registered as a student"})
		}
		actorID = student.ID
	default:
		return c.Status(403).JSON(fiber.Map{"error": "Access denied for your role"})
	}

	// Get stats
	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	stats, err := s.reportRepo.GetAchievementStats(ctx, actorID, role.Name, startDate, endDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to generate statistics",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    stats,
	})
}

// GetStudentReport godoc
// @Summary Get student achievement report
// @Description
// Mengambil laporan statistik prestasi untuk mahasiswa tertentu.
// Akses dibatasi berdasarkan role:
// - Admin: bebas
// - Dosen Wali: hanya mahasiswa bimbingannya
// - Mahasiswa: hanya data miliknya sendiri
//
// Mendukung filter rentang tanggal.
//
// @Tags Report
// @Security BearerAuth
// @Produce json
//
// @Param id path string true "Student UUID"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
//
// @Success 200 {object} map[string]interface{} "Student report statistics"
// @Failure 400 {object} map[string]string "Invalid student ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Access denied"
// @Failure 404 {object} map[string]string "Student not found"
// @Failure 500 {object} map[string]string "Failed to generate report"
//
// @Router /reports/student/{id} [get]
func (s *ReportService) GetStudentReport(c *fiber.Ctx) error {
	studentIDStr := c.Params("id")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid student ID"})
	}

	currentUser, ok := c.Locals("user").(*models.User)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
	}

	student, err := s.studentRepo.GetByID(studentID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to get student"})
	}
	if student == nil {
		return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
	}

	// Authorization
	role, err := s.roleRepo.GetByID(currentUser.RoleID)
	if err != nil || role == nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to verify user role"})
	}

	authorized := false
	switch role.Name {
	case "Admin":
		authorized = true
	case "Dosen Wali":
		lecturer, _ := s.lecturerRepo.GetByUserID(currentUser.ID)
		authorized = lecturer != nil && student.AdvisorID != nil && *student.AdvisorID == lecturer.ID
	case "Mahasiswa":
		authorized = student.UserID == currentUser.ID
	}

	if !authorized {
		return c.Status(403).JSON(fiber.Map{"error": "Access denied"})
	}

	// Parse dates
	var startDate, endDate *time.Time
	if startStr := c.Query("start_date"); startStr != "" {
		if parsed, err := time.Parse("2006-01-02", startStr); err == nil {
			startDate = &parsed
		}
	}
	if endStr := c.Query("end_date"); endStr != "" {
		if parsed, err := time.Parse("2006-01-02", endStr); err == nil {
			endDate = &parsed
		}
	}

	// Get stats
	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	stats, err := s.reportRepo.GetAchievementStats(ctx, studentID, "Mahasiswa", startDate, endDate)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to generate student statistics",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"student_id": student.ID,
			"statistics": stats,
		},
	})
}