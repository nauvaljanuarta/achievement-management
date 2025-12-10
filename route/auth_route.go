package route

import (
	"achievement-backend/middleware"
	"achievement-backend/app/repository"
	
	"github.com/gofiber/fiber/v2"
)

func setupAuthRoutes(
	router fiber.Router, 
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) {
	authRoutes := router.Group("/auth")
	
	authRoutes.Post("/login", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Login endpoint - akan diimplementasi nanti",
		})
	})
	
	authRoutes.Post("/refresh", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Refresh token endpoint - akan diimplementasi nanti",
		})
	})
	
	authRoutes.Post("/logout", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Logout endpoint - akan diimplementasi nanti",
		})
	})
	
	// Protected auth endpoint
	authRoutes.Get("/profile", 
		middleware.RequireAuth(userRepo),
		func(c *fiber.Ctx) error {
			userID, _ := c.Locals("user_id").(string)
			return c.JSON(fiber.Map{
				"message": "Profile endpoint",
				"user_id": userID,
			})
		},
	)
}