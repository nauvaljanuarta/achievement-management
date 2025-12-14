package service

import (
	"errors"
	"fmt"
	"time"

	"achievement-backend/app/models"
	"achievement-backend/app/repository"
	"achievement-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	studentRepo  repository.StudentRepository
	lecturerRepo repository.LecturerRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		studentRepo:  studentRepo,
		lecturerRepo: lecturerRepo,
	}
}

func (s *UserService) GetAll(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	
	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get active users with pagination
	users, total, err := s.userRepo.GetAll(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get users",
			"details": err.Error(),
		})
	}

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit // ceil division
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"data": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}

func (s *UserService) GetInactiveUsers(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Get inactive users
	users, total, err := s.userRepo.GetInactiveUsers(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get inactive users",
			"details": err.Error(),
		})
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"data": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}

func (s *UserService) GetByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Get active user only
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get user",
			"details": err.Error(),
		})
	}
	if user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found or inactive",
		})
	}

	return c.JSON(fiber.Map{
		"data": user,
	})
}

func (s *UserService) Create(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to validate role",
			"details": err.Error(),
		})
	}
	if role == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	existingUser, _ := s.userRepo.GetByUsername(req.Username)
	if existingUser != nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	existingUser, _ = s.userRepo.GetByEmail(req.Email)
	if existingUser != nil {
		return c.Status(409).JSON(fiber.Map{
			"error": "Email already exists",
		})
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to hash password",
			"details": err.Error(),
		})
	}

	user := &models.User{
		ID:           uuid.New(),
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		RoleID:       roleID,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	userID, err := s.userRepo.Create(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create user",
			"details": err.Error(),
		})
	}
	user.ID = userID

	if err := s.createUserProfile(user, &req, role.Name); err != nil {
		if deleteErr := s.userRepo.HardDelete(user.ID); deleteErr != nil {
			fmt.Printf("CRITICAL: Failed to rollback user creation: %v\n", deleteErr)
		}
		
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to create user profile",
			"details": err.Error(), 
		})
	}

	createdUser, err := s.userRepo.GetByID(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch created user",
			"details": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "User created successfully",
		"data":    createdUser,
	})
}

func (s *UserService) createUserProfile(user *models.User, req *models.CreateUserRequest, roleName string) error {
	switch roleName {
	case "Mahasiswa":
		if req.StudentID == nil || req.ProgramStudy == nil || req.AcademicYear == nil {
			return errors.New("student data incomplete: need studentId, programStudy, and academicYear")
		}

		student := models.Student{
			ID:           uuid.New(),
			UserID:       user.ID,
			StudentID:    *req.StudentID,
			ProgramStudy: *req.ProgramStudy,
			AcademicYear: *req.AcademicYear,
			AdvisorID:    nil,
			CreatedAt:    time.Now(),
		}

		if req.AdvisorID != nil && *req.AdvisorID != "" {
			advisorID, err := uuid.Parse(*req.AdvisorID)
			if err != nil {
					return fmt.Errorf("invalid advisor ID format: %w", err)
			}
			
			advisor, err := s.lecturerRepo.GetByID(advisorID)
			if err != nil {
					return fmt.Errorf("error checking advisor: %w", err)
			}
			if advisor == nil {
					return fmt.Errorf("advisor not found with ID: %s", advisorID)
			}
			
			student.AdvisorID = &advisorID
	}

		_, err := s.studentRepo.Create(student)
		return err

	case "Dosen Wali":
		if req.LecturerID == nil || req.Department == nil {
			return errors.New("lecturer data incomplete: need lecturerId and department")
		}

		lecturer := models.Lecturer{
			ID:         uuid.New(),
			UserID:     user.ID,
			LecturerID: *req.LecturerID,
			Department: *req.Department,
			CreatedAt:  time.Now(),
		}

		_, err := s.lecturerRepo.Create(lecturer)
		return err
	}

	return nil
}


func (s *UserService) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if user exists (active only)
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check user",
			"details": err.Error(),
		})
	}
	if user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found or inactive",
		})
	}

	// Validate email uniqueness if updating email
	if req.Email != nil {
		existingUser, _ := s.userRepo.GetByEmail(*req.Email)
		if existingUser != nil && existingUser.ID != id {
			return c.Status(409).JSON(fiber.Map{
				"error": "Email already used by another user",
			})
		}
	}

	// Update password if provided
	if req.Password != nil {
		hashedPassword, err := utils.HashPassword(*req.Password)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to hash password",
				"details": err.Error(),
			})
		}
		
		if err := s.userRepo.UpdatePassword(id, hashedPassword); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to update password",
				"details": err.Error(),
			})
		}
	}

	// Update other user data
	if err := s.userRepo.Update(id, &req); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update user",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
	})
}


func (s *UserService) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// Check if user exists (active only)
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check user",
			"details": err.Error(),
		})
	}
	if user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found or already inactive",
		})
	}

	// Soft delete user
	if err := s.userRepo.SoftDelete(id); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to delete user",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully (soft delete)",
	})
}

func (s *UserService) UpdateRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req struct {
		RoleID string `json:"roleId"`
	}
	
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid role ID",
		})
	}

	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check user",
			"details": err.Error(),
		})
	}
	if user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found or inactive",
		})
	}

	// Check if role exists
	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check role",
			"details": err.Error(),
		})
	}
	if role == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Role not found",
		})
	}

	// Update role
	roleIDStr := roleID.String()
	updateReq := &models.UpdateUserRequest{
		RoleID: &roleIDStr,
	}
	
	if err := s.userRepo.Update(id, updateReq); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update role",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User role updated successfully",
	})
}

func (s *UserService) SearchByName(c *fiber.Ctx) error {
	name := c.Query("name", "")
	if name == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name parameter is required",
		})
	}

	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := s.userRepo.SearchByName(name, page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to search users",
			"details": err.Error(),
		})
	}

	totalPages := (total + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	return c.JSON(fiber.Map{
		"data": users,
		"pagination": fiber.Map{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	})
}