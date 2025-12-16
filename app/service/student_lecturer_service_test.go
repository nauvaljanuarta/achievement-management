package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"achievement-backend/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- HELPER SETUP ---

func setupStudentLecturerApp(svc *StudentLecturerService, user *models.User) *fiber.App {
	app := fiber.New()

	// Middleware Mock: Inject user login ke Locals
	app.Use(func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("user", user)
			c.Locals("user_id", user.ID)
		}
		return c.Next()
	})

	// Register Routes sesuai Service
	app.Get("/students", svc.GetAllStudents)
	app.Get("/students/:id", svc.GetStudentByID)
	app.Put("/students/:id/advisor", svc.UpdateStudentAdvisor)
	app.Get("/students/:id/achievements", svc.GetStudentAchievements)
	app.Get("/lecturers/:id/advisees", svc.GetLecturerAdvisees)

	return app
}

// --- TEST CASES ---

func TestGetAllStudents_Success(t *testing.T) {
	// 1. Setup Data
	mockStudents := []models.StudentResponse{
		{ID: uuid.New(), FullName: "Mahasiswa A", StudentID: "101"},
		{ID: uuid.New(), FullName: "Mahasiswa B", StudentID: "102"},
	}

	// 2. Setup Service dengan Mock
	svc := &StudentLecturerService{
		studentRepo: &MockStudentRepository{
			GetWithUserDetailsFn: func(page, limit int) ([]models.StudentResponse, int, error) {
				return mockStudents, 2, nil
			},
		},
	}

	app := setupStudentLecturerApp(svc, nil) // Tidak butuh login untuk route ini (sesuai kode)

	// 3. Request
	req := httptest.NewRequest("GET", "/students?page=1&limit=10", nil)
	resp, _ := app.Test(req, -1)

	// 4. Assert
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)

	// Cek pagination total
	pagination := responseBody["pagination"].(map[string]interface{})
	if pagination["total"].(float64) != 2 {
		t.Errorf("Expected total 2, got %v", pagination["total"])
	}

	// Cek data
	data := responseBody["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 students, got %d", len(data))
	}
}

func TestUpdateStudentAdvisor_Assign(t *testing.T) {
	studentID := uuid.New()
	lecturerID := uuid.New()
	userID := uuid.New()

	// Setup Mock
	svc := &StudentLecturerService{
		studentRepo: &MockStudentRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: userID}, nil
			},
			UpdateAdvisorFn: func(sID, aID uuid.UUID) error {
				if sID == studentID && aID == lecturerID {
					return nil // Sukses
				}
				return fiber.ErrBadRequest
			},
		},
		lecturerRepo: &MockLecturerRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Lecturer, error) {
				return &models.Lecturer{ID: lecturerID}, nil
			},
		},
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return &models.User{FullName: "Dosen Keren"}, nil
			},
		},
	}

	app := setupStudentLecturerApp(svc, nil)

	// Payload Assign Advisor
	payload := map[string]string{
		"advisor_id": lecturerID.String(),
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/students/"+studentID.String()+"/advisor", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)
	data := responseBody["data"].(map[string]interface{})

	// Cek action response
	if data["action"] != "assigned" {
		t.Errorf("Expected action 'assigned', got %s", data["action"])
	}
	advisor := data["advisor"].(map[string]interface{})
	if advisor["name"] != "Dosen Keren" {
		t.Errorf("Expected advisor name 'Dosen Keren', got %s", advisor["name"])
	}
}

func TestGetStudentAchievements_Success(t *testing.T) {
	studentID := uuid.New()
	userID := uuid.New() // ID User yang login & pemilik student sama
	mongoID := primitive.NewObjectID()

	mockUser := &models.User{ID: userID, FullName: "Mhs Berprestasi"}

	svc := &StudentLecturerService{
		studentRepo: &MockStudentRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: userID}, nil
			},
		},
		achievementRefRepo: &MockAchievementReferenceRepository{
			FindByStudentIDFn: func(ctx context.Context, sID uuid.UUID, s string, p, l int) ([]*models.AchievementReference, int, error) {
				return []*models.AchievementReference{
					{
						ID:                 uuid.New(),
						MongoAchievementID: mongoID.Hex(),
						Status:             "verified",
						SubmittedAt:        &time.Time{},
					},
				}, 1, nil
			},
		},
		achievementRepo: &MockAchievementRepository{
			FindByIDFn: func(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error) {
				if id == mongoID {
					return &models.Achievement{
						Title:           "Juara 1 Lomba",
						Points:          100,
						AchievementType: "Lomba",
					}, nil
				}
				return nil, nil
			},
		},
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return mockUser, nil
			},
		},
	}

	// Login sebagai pemilik student (agar lolos permission check)
	app := setupStudentLecturerApp(svc, mockUser)

	req := httptest.NewRequest("GET", "/students/"+studentID.String()+"/achievements", nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)
	data := responseBody["data"].(map[string]interface{})

	achievements := data["achievements"].([]interface{})
	if len(achievements) == 0 {
		t.Error("Should return achievements")
	}

	firstAch := achievements[0].(map[string]interface{})
	if firstAch["title"] != "Juara 1 Lomba" {
		t.Errorf("Expected title 'Juara 1 Lomba', got %s", firstAch["title"])
	}
}

func TestGetStudentByID_AccessDenied(t *testing.T) {
	// Skenario: User A mencoba akses profil Student milik User B
	// User A bukan Admin dan bukan Dosen Wali dari Student B

	ownerUserID := uuid.New()
	attackerUserID := uuid.New()
	studentID := uuid.New()

	attackerUser := &models.User{
		ID:       attackerUserID,
		RoleID:   uuid.New(),
		FullName: "Attacker",
	}

	svc := &StudentLecturerService{
		studentRepo: &MockStudentRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: ownerUserID}, nil
			},
		},
		lecturerRepo: &MockLecturerRepository{
			GetByUserIDFn: func(uid uuid.UUID) (*models.Lecturer, error) {
				return nil, nil // Attacker bukan dosen
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil // Role Attacker hanya mahasiswa
			},
		},
	}

	// Login sebagai Attacker
	app := setupStudentLecturerApp(svc, attackerUser)

	req := httptest.NewRequest("GET", "/students/"+studentID.String(), nil)
	resp, _ := app.Test(req, -1)

	if resp.StatusCode != 403 {
		t.Errorf("Expected 403 Forbidden, got %d", resp.StatusCode)
	}

	var responseBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)
	if responseBody["error"] != "Access denied" {
		t.Errorf("Expected 'Access denied', got %s", responseBody["error"])
	}
}