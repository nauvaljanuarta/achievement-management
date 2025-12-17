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
	studentRepo  repository.StudentRepository  
	lecturerRepo repository.LecturerRepository 
}

func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	studentRepo repository.StudentRepository,   
	lecturerRepo repository.LecturerRepository, ) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		studentRepo:  studentRepo,   
		lecturerRepo: lecturerRepo,  
	}
}

// Login godoc
// @Summary Login user
// @Description Login menggunakan username atau email dan password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login Request"
// @Success 200 {object} models.LoginResponse "Login berhasil"
// @Failure 400 {object} object{error=string} "Invalid request body"
// @Failure 401 {object} object{error=string} "Invalid credentials"
// @Failure 403 {object} object{error=string} "Account is inactive"
// @Failure 500 {object} object{error=string,details=string} "Server error"
// @Router /auth/login [post]
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


// RefreshToken godoc
// @Summary Refresh Access Token
// @Description Generates a new access token and refresh token using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body object{refreshToken=string} true "Refresh Token Request"
// @Success 200 {object} object{token=string,refreshToken=string,user=object,permissions=[]string,roleName=string} "Successfully refreshed tokens"
// @Failure 400 {object} object{error=string} "Invalid request body or missing refresh token"
// @Failure 401 {object} object{error=string,details=string} "Invalid/expired refresh token or user not found"
// @Failure 403 {object} object{error=string} "Account is inactive"
// @Failure 500 {object} object{error=string,details=string} "Server error"
// @Router /auth/refresh [post]
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

// Logout godoc
// @Summary Logout user
// @Description Logout user (client-side token invalidation)
// @Tags Authentication
// @Produce json
// @Success 200 {object} object{message=string,note=string} "Logout berhasil"
// @Security BearerAuth
// @Router /auth/logout [post]
func (s *AuthService) Logout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
		"note":    "Token Deleted with Exp Date",
	})
}

// Profile godoc
// @Summary Get user profile
// @Description Mengambil data profil user beserta role, permissions, dan data tambahan (mahasiswa/dosen wali)
// @Tags Authentication
// @Produce json
// @Success 200 {object} object{data=object} "Profil user"
// @Failure 401 {object} object{error=string} "User not authenticated"
// @Failure 404 {object} object{error=string} "User not found"
// @Failure 500 {object} object{error=string,details=string} "Server error"
// @Security BearerAuth
// @Router /auth/profile [get]
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


// PERBAIKAN: getStudentProfile sekarang menggunakan repository
func (s *AuthService) getStudentProfile(userID uuid.UUID) (*models.Student, error) {
	student, err := s.studentRepo.GetByUserID(userID)
	if err != nil {
		fmt.Printf("Error getting student profile for user %s: %v\n", userID, err)
		return nil, err
	}
	return student, nil
}

// PERBAIKAN: getLecturerProfile sekarang menggunakan repository
func (s *AuthService) getLecturerProfile(userID uuid.UUID) (*models.Lecturer, error) {
	lecturer, err := s.lecturerRepo.GetByUserID(userID)
	if err != nil {
		fmt.Printf("Error getting lecturer profile for user %s: %v\n", userID, err)
		return nil, err
	}
	return lecturer, nil
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