package route

import (
	// "achievement-backend/middleware"
	"achievement-backend/app/service"
	"achievement-backend/app/repository"

	"github.com/gofiber/fiber/v2"
)

func setupUserRoutes(
	router fiber.Router, 
	userService *service.UserService,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) {
	userRoutes := router.Group("/users")
	
	protectedUserRoutes := userRoutes.Group("",)
	// GET /api/v1/users - Get all active users
	userRoutes.Get("/", userService.GetAll)
	// GET /api/v1/users/:id - Get user by ID
	userRoutes.Get("/:id", userService.GetByID)
	// GET /api/v1/users/search - Search users by name
	userRoutes.Get("/search", userService.SearchByName)
	// ADMIN ONLY ROUTES
	// POST /api/v1/users - Create user (admin only)
	protectedUserRoutes.Post("/", userService.Create)
	// PUT /api/v1/users/:id - Update user (admin only)
	protectedUserRoutes.Put("/:id", userService.Update)
	// DELETE /api/v1/users/:id - Delete user (admin only)
	protectedUserRoutes.Delete("/:id", userService.Delete)
	// PUT /api/v1/users/:id/role - Update user role (admin only)
	protectedUserRoutes.Put("/:id/role", userService.UpdateRole)
	// GET /api/v1/users/inactive - Get inactive users (admin only)
	protectedUserRoutes.Get("/inactive", userService.GetInactiveUsers)
}