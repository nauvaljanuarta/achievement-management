package route

import (
	"achievement-backend/app/repository"
	"achievement-backend/app/service"
	"achievement-backend/middleware"

	"github.com/gofiber/fiber/v2"
)

func setupUserRoutes(
	router fiber.Router, 
	userService *service.UserService,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
) {
	userRoutes := router.Group("/users")
	
	protectedUserRoutes := userRoutes.Group("",middleware.RequireAuth(userRepo),middleware.AdminOnly(roleRepo),)
	userRoutes.Get("/", userService.GetAll)
	userRoutes.Get("/:id", userService.GetByID)
	// admin
	protectedUserRoutes.Post("/", userService.Create)
	protectedUserRoutes.Put("/:id", userService.Update)
	protectedUserRoutes.Delete("/:id", userService.Delete)
	protectedUserRoutes.Put("/:id/role", userService.UpdateRole)
}