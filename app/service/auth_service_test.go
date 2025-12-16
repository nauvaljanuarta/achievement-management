package service

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"achievement-backend/app/models"
	
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt" 
)

func setupAuthApp(svc *AuthService, user *models.User) *fiber.App {
	app := fiber.New()

	// Middleware Mock untuk route yang butuh login (Profile & Change Password)
	app.Use(func(c *fiber.Ctx) error {
		if user != nil {
			c.Locals("user", user)
			c.Locals("user_id", user.ID)
		}
		return c.Next()
	})

	// Routes
	app.Post("/auth/login", svc.Login)
	app.Get("/auth/profile", svc.Profile)

	return app
}

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes)
}


func TestLogin_Success(t *testing.T) {
	// 1. Setup Data Dummy
	userID := uuid.New()
	roleID := uuid.New()
	passwordRaw := "password123"
	passwordHashed := hashPassword(passwordRaw) // Hash beneran biar lolos validasi

	mockUser := &models.User{
		ID:           userID,
		Username:     "mahasiswa1",
		Email:        "mhs@univ.ac.id",
		PasswordHash: passwordHashed,
		RoleID:       roleID,
		IsActive:     true,
	}

	// 2. Setup Mock Service
	svc := &AuthService{
		userRepo: &MockUserRepository{
			GetByUsernameFn: func(username string) (*models.User, error) {
				if username == "mahasiswa1" {
					return mockUser, nil
				}
				return nil, nil
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil
			},
			GetPermissionNamesByRoleIDFn: func(roleID uuid.UUID) ([]string, error) {
				return []string{"achievement.create", "achievement.read"}, nil
			},
		},
	}

	app := setupAuthApp(svc, nil) // User nil karena belum login

	// 3. Request Body
	payload := map[string]string{
		"username": "mahasiswa1",
		"password": "password123", // Password yang cocok
	}
	body, _ := json.Marshal(payload)

	// 4. Execute
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	// 5. Assert
	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Cek apakah token ada di response
	var responseBody map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&responseBody)

	if responseBody["token"] == nil {
		t.Error("Token should be present in response")
	}
}


func TestLogin_WrongPassword(t *testing.T) {
	passwordHashed := hashPassword("passwordBenar")

	svc := &AuthService{
		userRepo: &MockUserRepository{
			GetByUsernameFn: func(username string) (*models.User, error) {
				return &models.User{
					Username:     "mahasiswa1",
					PasswordHash: passwordHashed,
					IsActive:     true,
				}, nil
			},
		},
	}

	app := setupAuthApp(svc, nil)

	payload := map[string]string{
		"username": "mahasiswa1",
		"password": "passwordSalah", // Password beda
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode != 401 {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
}


func TestProfile_Student(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()

	mockUser := &models.User{ID: userID, RoleID: roleID, FullName: "Budi Santoso"}

	svc := &AuthService{
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return mockUser, nil
			},
		},
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{Name: "Mahasiswa"}, nil
			},
			GetPermissionNamesByRoleIDFn: func(roleID uuid.UUID) ([]string, error) {
				return []string{}, nil
			},
		},
		studentRepo: &MockStudentRepository{
			GetByUserIDFn: func(id uuid.UUID) (*models.Student, error) {
				return &models.Student{ID: uuid.New(), StudentID: "0812345", ProgramStudy: "Informatika"}, nil
			},
		},
	}

	// Inject User via Middleware
	app := setupAuthApp(svc, mockUser)

	req := httptest.NewRequest("GET", "/auth/profile", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", resp.StatusCode)
	}

	// Validasi Response ada data studentProfile
	var responseBody map[string]map[string]interface{} // structure: { "data": { ... } }
	json.NewDecoder(resp.Body).Decode(&responseBody)

	data := responseBody["data"]
	if data["studentProfile"] == nil {
		t.Error("Student Profile should be present for Mahasiswa role")
	}
}
