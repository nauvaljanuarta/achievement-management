package middleware

import (
	"achievement-backend/utils"
	"strings"
	"github.com/gofiber/fiber/v2"
)

func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "missing or invalid token",
		})
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid authorization header",
		})
	}
	tokenString := parts[1]

	claims, err := utils.ValidateToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "invalid or expired token",
		})
	}

	c.Locals("user_id", claims.UserID)
	c.Locals("email", claims.Email)
	c.Locals("role", claims.Role)

	return c.Next()
}

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {
			role, ok := c.Locals("role").(string)
			if !ok || role != "admin" {
					return c.Status(403).JSON(fiber.Map{
							"error": "Akses ditolak. Hanya admin yang diizinkan",
					})
			}
			return c.Next()
	}
}
