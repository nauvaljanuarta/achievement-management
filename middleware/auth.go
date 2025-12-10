package middleware

import (
	"strings"

	"achievement-backend/app/repository"
	"achievement-backend/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RequireAuth(userRepo repository.UserRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "missing authorization header",
			})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "invalid authorization format",
			})
		}

		tokenString := parts[1]

		// Validate token
		claims, err := utils.ValidateToken(tokenString)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "invalid or expired token",
				"error":   err.Error(),
			})
		}

		// Verify user exists and is active
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "invalid user in token",
			})
		}

		user, err := userRepo.GetByID(userID)
		if err != nil || user == nil || !user.IsActive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"message": "user not found or inactive",
			})
		}

		c.Locals("user_id", user.ID)
		c.Locals("user", user)
		c.Locals("role_id", user.RoleID)
		c.Locals("permissions", claims.Permissions) // Dari JWT claims

		return c.Next()
	}
}

func RequirePermission(permission string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		permissions, ok := c.Locals("permissions").([]string)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "no permissions found",
			})
		}

		hasPermission := false
		for _, p := range permissions {
			if p == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "insufficient permissions",
				"required": permission,
			})
		}

		return c.Next()
	}
}

// AdminOnly - Hanya untuk role Admin
func AdminOnly(roleRepo repository.RoleRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID, ok := c.Locals("role_id").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			})
		}

		role, err := roleRepo.GetByID(roleID)
		if err != nil || role == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "role not found",
			})
		}

		if role.Name != "Admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "admin access required",
			})
		}

		return c.Next()
	}
}

func RequireRole(roleName string, roleRepo repository.RoleRepository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID, ok := c.Locals("role_id").(uuid.UUID)
		if !ok {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			})
		}

		role, err := roleRepo.GetByID(roleID)
		if err != nil || role == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "role not found",
			})
		}

		if role.Name != roleName {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": roleName + " access required",
				"required_role": roleName,
				"user_role": role.Name,
			})
		}

		return c.Next()
	}
}