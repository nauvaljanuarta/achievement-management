package service

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"
	"io"

	"achievement-backend/app/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func setupUserApp(svc *UserService) *fiber.App {
	app := fiber.New()
	app.Post("/users", svc.Create)
	app.Get("/users", svc.GetAll)
	app.Get("/users/:id", svc.GetByID)
	app.Put("/users/:id", svc.Update)
	app.Delete("/users/:id", svc.Delete)
	return app
}

func TestCreateUser_Success(t *testing.T) {
	roleID := uuid.New()
	userID := uuid.New()

	svc := &UserService{
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{ID: roleID, Name: "Mahasiswa"}, nil
			},
		},
		userRepo: &MockUserRepository{
			// GetByUsernameFn harus return nil (user belum ada)
			GetByUsernameFn: func(username string) (*models.User, error) {
				return nil, nil
			},
			// GetByEmailFn harus return nil (email belum ada)
			GetByEmailFn: func(email string) (*models.User, error) {
				return nil, nil
			},
			CreateFn: func(user *models.User) (uuid.UUID, error) {
				return userID, nil
			},
			// GetByIDFn dipanggil di akhir fungsi Create untuk return response
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return &models.User{ID: userID, Username: "newuser"}, nil
			},
		},
		studentRepo: &MockStudentRepository{
			CreateFn: func(s models.Student) (uuid.UUID, error) {
				return uuid.New(), nil
			},
		},
		lecturerRepo: &MockLecturerRepository{
        },
	}

	app := setupUserApp(svc)

	studentID := "123456"
	programStudy := "Informatika"
	academicYear := "2023"

	payload := models.CreateUserRequest{
		Username:     "newuser",
		Email:        "new@example.com",
		Password:     "password123",
		FullName:     "New User",
		RoleID:       roleID.String(),
		StudentID:    &studentID,
		ProgramStudy: &programStudy,
		AcademicYear: &academicYear,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
			t.Fatalf("Request error: %v", err)
	}

	if resp.StatusCode != 201 {
			bodyBytes, _ := io.ReadAll(resp.Body)
			t.Errorf("Expected 201, got %d. Response: %s", resp.StatusCode, string(bodyBytes))
			return
	}
}

func TestCreateUser_Duplicate(t *testing.T) {
	roleID := uuid.New()

	svc := &UserService{
		roleRepo: &MockRoleRepository{
			GetByIDFn: func(id uuid.UUID) (*models.Role, error) {
				return &models.Role{ID: roleID, Name: "Admin"}, nil
			},
		},
		userRepo: &MockUserRepository{
			GetByUsernameFn: func(username string) (*models.User, error) {
				return &models.User{Username: "existing"}, nil
			},
		},
	}

	app := setupUserApp(svc)

	payload := models.CreateUserRequest{
		Username: "existing",
		Email:    "test@example.com",
		Password: "password123",
		RoleID:   roleID.String(),
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	if resp.StatusCode != 409 {
		t.Errorf("Expected 409, got %d", resp.StatusCode)
	}
}

func TestGetUserByID_Success(t *testing.T) {
	userID := uuid.New()

	svc := &UserService{
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return &models.User{ID: userID, Username: "testuser", IsActive: true}, nil
			},
		},
	}

	app := setupUserApp(svc)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestGetAllUsers_Success(t *testing.T) {
	svc := &UserService{
		userRepo: &MockUserRepository{
			GetAllFn: func(page, limit int) ([]models.User, int, error) {
				return []models.User{
					{Username: "user1"},
					{Username: "user2"},
				}, 2, nil
			},
		},
	}

	app := setupUserApp(svc)

	req := httptest.NewRequest("GET", "/users?page=1&limit=10", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	userID := uuid.New()

	svc := &UserService{
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return &models.User{ID: userID, Username: "oldname", IsActive: true}, nil
			},
			GetByEmailFn: func(email string) (*models.User, error) {
				return nil, nil
			},
			UpdateFn: func(id uuid.UUID, req *models.UpdateUserRequest) error {
				return nil
			},
		},
	}

	app := setupUserApp(svc)

	newEmail := "updated@example.com"
	payload := models.UpdateUserRequest{
		Email: &newEmail,
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("PUT", "/users/"+userID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	userID := uuid.New()

	svc := &UserService{
		userRepo: &MockUserRepository{
			GetByIDFn: func(id uuid.UUID) (*models.User, error) {
				return &models.User{ID: userID, IsActive: true}, nil
			},
			SoftDeleteFn: func(id uuid.UUID) error {
				return nil
			},
		},
	}

	app := setupUserApp(svc)

	req := httptest.NewRequest("DELETE", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	if resp.StatusCode != 200 {
		t.Errorf("Expected 200, got %d", resp.StatusCode)
	}
}