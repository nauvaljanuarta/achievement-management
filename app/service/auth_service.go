package service

import (
	"fmt"

	"achievement-backend/app/models"
	"achievement-backend/app/repository"
	"achievement-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo repository.UserRepository
	roleRepo repository.RoleRepository
}

func NewAuthService(userRepo repository.UserRepository, roleRepo repository.RoleRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var user *models.User
	var err error

	user, err = s.userRepo.GetByUsername(req.Username)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error checking user",
			"details": err.Error(),
		})
	}

	if user == nil {
		user, err = s.userRepo.GetByEmail(req.Email)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Error checking user",
				"details": err.Error(),
			})
		}
	}

	if user == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{
			"error": "Account is inactive",
		})
	}

	role, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error getting role information",
			"details": err.Error(),
		})
	}

	permissions, err := s.roleRepo.GetPermissionNamesByRoleID(user.RoleID)
	if err != nil {
		permissions = []string{}
		fmt.Printf("Warning: Failed to get permissions for role %s: %v\n", role.Name, err)
	}

	token, err := utils.GenerateToken(user, permissions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate token",
			"details": err.Error(),
		})
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate refresh token",
			"details": err.Error(),
		})
	}

	response := models.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         *user,
		Permissions:  permissions,
		RoleName:     role.Name,
	}

	return c.JSON(response)
}


func (s *AuthService) RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.RefreshToken == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Refresh token is required",
		})
	}

	claims, err := utils.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid or expired refresh token",
			"details": err.Error(),
		})
	}

	userID, err := uuid.Parse(claims.Subject)  
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid user ID in token",
			"details": err.Error(),
		})
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error checking user",
			"details": err.Error(),
		})
	}

	if user == nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if !user.IsActive {
		return c.Status(403).JSON(fiber.Map{
			"error": "Account is inactive",
		})
	}

	role, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error getting role information",
			"details": err.Error(),
		})
	}

	permissions, err := s.roleRepo.GetPermissionNamesByRoleID(user.RoleID)
	if err != nil {
		permissions = []string{}
	}

	newToken, err := utils.GenerateToken(user, permissions)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate new token",
			"details": err.Error(),
		})
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate new refresh token",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token":        newToken,
		"refreshToken": newRefreshToken,
		"user":         user,
		"permissions":  permissions,
		"roleName":     role.Name,
	})
}


func (s *AuthService) Logout(c *fiber.Ctx) error {
	// Simple logout - client side delete tokens
	// Untuk production, bisa implement token blacklist atau revoke di sini

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
		"note":    "Please delete tokens on client side",
	})
}


func (s *AuthService) Profile(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error fetching user",
			"details": err.Error(),
		})
	}

	if user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Get role
	role, err := s.roleRepo.GetByID(user.RoleID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Error fetching role",
			"details": err.Error(),
		})
	}

	// Get permissions
	permissions, err := s.roleRepo.GetPermissionNamesByRoleID(user.RoleID)
	if err != nil {
		permissions = []string{}
	}

	profileData := fiber.Map{
		"user":        user,
		"role":        role,
		"permissions": permissions,
	}

	if role.Name == "Mahasiswa" {
		student, err := s.getStudentProfile(user.ID)
		if err == nil && student != nil {
			profileData["studentProfile"] = student
		}
	}

	if role.Name == "Dosen Wali" {
		lecturer, err := s.getLecturerProfile(user.ID)
		if err == nil && lecturer != nil {
			profileData["lecturerProfile"] = lecturer
		}
	}

	return c.JSON(fiber.Map{
		"data": profileData,
	})
}


func (s *AuthService) getStudentProfile(userID uuid.UUID) (*models.Student, error) {
	// Butuh StudentRepository di AuthService
	// Untuk sekarang return nil, nanti bisa di-extend
	return nil, nil
}

func (s *AuthService) getLecturerProfile(userID uuid.UUID) (*models.Lecturer, error) {
	// Butuh LecturerRepository di AuthService
	return nil, nil
}

func (s *AuthService) ChangePassword(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(401).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validasi
	if req.NewPassword != req.ConfirmPassword {
		return c.Status(400).JSON(fiber.Map{
			"error": "New password and confirmation do not match",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(400).JSON(fiber.Map{
			"error": "New password must be at least 6 characters",
		})
	}

	// Get user
	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Verify current password
	if !utils.CheckPasswordHash(req.CurrentPassword, user.PasswordHash) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Current password is incorrect",
		})
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to hash password",
			"details": err.Error(),
		})
	}

	// Update password
	if err := s.userRepo.UpdatePassword(user.ID, hashedPassword); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update password",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password updated successfully",
	})
}