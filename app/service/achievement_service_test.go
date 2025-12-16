package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"achievement-backend/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/* =====================================================
   HELPER: Setup App
   ===================================================== */
func setupTestApp(svc *AchievementService, user *models.User) *fiber.App {
	app := fiber.New()

	// Middleware Mock: Inject user login
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user", user)
		c.Locals("user_id", user.ID)
		return c.Next()
	})

	// Routes Penting
	app.Post("/achievements", svc.CreateAchievement)
	app.Get("/achievements/:id", svc.GetAchievementByID)
	app.Put("/achievements/:id", svc.UpdateAchievement) // Route Update
	app.Post("/submit/:id", svc.SubmitAchievement)      // Route Submit

	return app
}

/* =====================================================
   1. TEST CREATE (Flow Utama)
   ===================================================== */
func TestCreateAchievement_Success(t *testing.T) {
	// Setup ID
	userID := uuid.New()
	studentID := uuid.New()

	// Setup Service dengan Mock
	svc := &AchievementService{
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil
			},
		},
		studentRepo: &MockStudentRepository{
			// PERBAIKAN: Menggunakan GetByUserIDFn (sesuai mocks.go)
			GetByUserIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: userID}, nil
			},
		},
		achievementRepo: &MockAchievementRepository{
			CreateFn: func(ctx context.Context, a *models.Achievement) (primitive.ObjectID, error) {
				return primitive.NewObjectID(), nil // Sukses simpan ke Mongo
			},
			DeleteFn: func(ctx context.Context, id primitive.ObjectID) error { return nil },
		},
		achievementRefRepo: &MockAchievementReferenceRepository{
			CreateFn: func(ctx context.Context, ref *models.AchievementReference) error {
				return nil // Sukses simpan ke Postgres
			},
		},
	}

	app := setupTestApp(svc, &models.User{ID: userID, RoleID: uuid.New()})

	// Request Body
	payload := map[string]interface{}{
		"title":            "Juara 1 Hackathon",
		"achievement_type": "competition",
		"description":      "Menang juara 1",
		"points":           100,
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != 201 {
		t.Errorf("Expected 201 Created, got %d", resp.StatusCode)
	}
}

/* =====================================================
   2. TEST GET (Logic Authorization)
   ===================================================== */
func TestGetAchievement_Success(t *testing.T) {
	refID := uuid.New()
	userID := uuid.New()
	studentID := uuid.New()

	svc := &AchievementService{
		// Mock Ref: Ditemukan & status Draft
		achievementRefRepo: &MockAchievementReferenceRepository{
			FindByIDFn: func(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error) {
				return &models.AchievementReference{
					ID:                 refID,
					MongoAchievementID: primitive.NewObjectID().Hex(),
					StudentID:          studentID, // PENTING: Milik user yg login
					Status:             "draft",
				}, nil
			},
		},
		// Mock Mongo: Detail ditemukan
		achievementRepo: &MockAchievementRepository{
			FindByIDFn: func(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error) {
				return &models.Achievement{Title: "Juara 1", AchievementType: "competition"}, nil
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil
			},
		},
		studentRepo: &MockStudentRepository{
			// PERBAIKAN: Menggunakan GetByUserIDFn untuk cek ownership
			GetByUserIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: userID}, nil
			},
			// GetByIDFn dipakai untuk ambil detail response
			GetByIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID, UserID: userID}, nil
			},
		},
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return &models.User{FullName: "Budi"}, nil
			},
		},
	}

	app := setupTestApp(svc, &models.User{ID: userID, RoleID: uuid.New()})

	req := httptest.NewRequest("GET", "/achievements/"+refID.String(), nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}
}

/* =====================================================
   3. TEST UPDATE (Logic Cek Status Draft)
   ===================================================== */
func TestUpdateAchievement_Success(t *testing.T) {
	refID := uuid.New()
	userID := uuid.New()
	studentID := uuid.New()

	svc := &AchievementService{
		achievementRefRepo: &MockAchievementReferenceRepository{
			FindByIDFn: func(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error) {
				return &models.AchievementReference{
					ID:                 refID,
					MongoAchievementID: primitive.NewObjectID().Hex(),
					StudentID:          studentID,
					Status:             "draft", // PENTING: Harus draft biar bisa diedit
				}, nil
			},
		},
		achievementRepo: &MockAchievementRepository{
			FindByIDFn: func(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error) {
				return &models.Achievement{Title: "Judul Lama"}, nil
			},
			UpdateFn: func(ctx context.Context, id primitive.ObjectID, a *models.Achievement) error {
				return nil // Sukses update
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil
			},
		},
		studentRepo: &MockStudentRepository{
			// PERBAIKAN: Menggunakan GetByUserIDFn
			GetByUserIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID}, nil
			},
		},
	}

	app := setupTestApp(svc, &models.User{ID: userID, RoleID: uuid.New()})

	// Kita update judulnya
	payload := map[string]interface{}{"title": "Judul Baru Revisi"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("PUT", "/achievements/"+refID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK (Update Success), got %d", resp.StatusCode)
	}
}

/* =====================================================
   4. TEST SUBMIT (Logic Transisi Status)
   ===================================================== */
func TestSubmitAchievement_Success(t *testing.T) {
	refID := uuid.New()
	userID := uuid.New()
	studentID := uuid.New()

	svc := &AchievementService{
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil
			},
		},
		studentRepo: &MockStudentRepository{
			// PERBAIKAN: Menggunakan GetByUserIDFn
			GetByUserIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: studentID}, nil
			},
		},
		achievementRefRepo: &MockAchievementReferenceRepository{
			FindByIDFn: func(ctx context.Context, id uuid.UUID) (*models.AchievementReference, error) {
				return &models.AchievementReference{
					ID:        refID,
					StudentID: studentID,
					Status:    "draft", // Awalnya Draft
				}, nil
			},
			SubmitForVerificationFn: func(ctx context.Context, id uuid.UUID) error {
				return nil // Sukses ubah status jadi submitted
			},
		},
	}

	app := setupTestApp(svc, &models.User{ID: userID, RoleID: uuid.New()})

	req := httptest.NewRequest("POST", "/submit/"+refID.String(), nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK (Submitted), got %d", resp.StatusCode)
	}
}