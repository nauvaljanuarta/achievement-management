package service

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"achievement-backend/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// --- HELPER SETUP ---
func setupReportApp(svc *ReportService, user *models.User) *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("user", user)
			c.Locals("user_id", user.ID)
		}
		return c.Next()
	})

	app.Get("/reports/statistics", svc.GetStatistics)
	app.Get("/reports/students/:id", svc.GetStudentReport)

	return app
}

// --- TEST CASES ---

func TestGetStatistics_Mahasiswa(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	studentID := uuid.New()

	mockUser := &models.User{ID: userID, RoleID: roleID, FullName: "Mhs Test"}

	svc := &ReportService{
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{ID: roleID, Name: "Mahasiswa"}, nil
			},
		},
		studentRepo: &MockStudentRepository{
			GetByUserIDFn: func(uid uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: userID}, nil
			},
		},
		reportRepo: &MockReportRepository{
			GetAchievementStatsFn: func(ctx context.Context, actorID uuid.UUID, roleName string, start, end *time.Time) (*models.AchievementStats, error) {
				if actorID != studentID {
					t.Errorf("Expected actorID %s, got %s", studentID, actorID)
				}
				
				// RETURN SESUAI MODEL KAMU
				return &models.AchievementStats{
					ByType: map[string]int{
						"Kompetisi": 5,
						"Organisasi": 3,
					},
					ByPeriod: map[string]int{"2023": 8},
				}, nil
			},
		},
	}

	app := setupReportApp(svc, mockUser)

	req := httptest.NewRequest("GET", "/reports/statistics?start_date=2023-01-01", nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)
	data := responseBody["data"].(map[string]interface{})

	// Cek Data map[string]int "by_type"
	byType := data["by_type"].(map[string]interface{})
	
	// JSON number di Go defaultnya float64 saat di-decode ke interface{}
	if byType["Kompetisi"].(float64) != 5 {
		t.Errorf("Expected 5 Kompetisi, got %v", byType["Kompetisi"])
	}
}

func TestGetStatistics_DosenWali(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	lecturerID := uuid.New()

	mockUser := &models.User{ID: userID, RoleID: roleID}

	svc := &ReportService{
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{ID: roleID, Name: "Dosen Wali"}, nil
			},
		},
		lecturerRepo: &MockLecturerRepository{
			GetByUserIDFn: func(uid uuid.UUID) (*models.Lecturer, error) {
				return &models.Lecturer{ID: lecturerID, UserID: userID}, nil
			},
		},
		reportRepo: &MockReportRepository{
			GetAchievementStatsFn: func(ctx context.Context, actorID uuid.UUID, roleName string, start, end *time.Time) (*models.AchievementStats, error) {
				// RETURN SESUAI MODEL KAMU (Top Students)
				return &models.AchievementStats{
					TopStudents: []models.StudentAchievementSum{
						{StudentName: "Budi", TotalPoints: 100},
					},
				}, nil
			},
		},
	}

	app := setupReportApp(svc, mockUser)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
	
	var responseBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)
	data := responseBody["data"].(map[string]interface{})
	
	// Cek Top Students
	topStudents := data["top_students"].([]interface{})
	firstStudent := topStudents[0].(map[string]interface{})
	
	if firstStudent["student_name"] != "Budi" {
		t.Errorf("Expected student name Budi, got %v", firstStudent["student_name"])
	}
}

func TestGetStudentReport_Success_Advisor(t *testing.T) {
	dosenUserID := uuid.New()
	lecturerID := uuid.New()
	studentID := uuid.New()
	
	mockDosenUser := &models.User{ID: dosenUserID}

	svc := &ReportService{
		studentRepo: &MockStudentRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, AdvisorID: &lecturerID}, nil
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Dosen Wali"}, nil
			},
		},
		lecturerRepo: &MockLecturerRepository{
			GetByUserIDFn: func(uid uuid.UUID) (*models.Lecturer, error) {
				return &models.Lecturer{ID: lecturerID, UserID: dosenUserID}, nil
			},
		},
		reportRepo: &MockReportRepository{
			GetAchievementStatsFn: func(ctx context.Context, actorID uuid.UUID, roleName string, start, end *time.Time) (*models.AchievementStats, error) {
				// Return kosong tapi valid sesuai struct
				return &models.AchievementStats{
					ByType: map[string]int{},
				}, nil
			},
		},
	}

	app := setupReportApp(svc, mockDosenUser)

	req := httptest.NewRequest("GET", "/reports/students/"+studentID.String(), nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestGetStudentReport_AccessDenied(t *testing.T) {
	attackerID := uuid.New()
	targetStudentID := uuid.New()
	ownerUserID := uuid.New()

	mockAttacker := &models.User{ID: attackerID}

	svc := &ReportService{
		studentRepo: &MockStudentRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: targetStudentID, UserID: ownerUserID}, nil
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil
			},
		},
		reportRepo: &MockReportRepository{},
	}

	app := setupReportApp(svc, mockAttacker)

	req := httptest.NewRequest("GET", "/reports/students/"+targetStudentID.String(), nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 403 {
		t.Errorf("Expected 403 Forbidden, got %d", resp.StatusCode)
	}
}