package route

import (
	"achievement-backend/middleware"
	"achievement-backend/app/repository"
	"achievement-backend/app/service"
	
	"github.com/gofiber/fiber/v2"
)

func setupAuthRoutes(
	router fiber.Router,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) {
	authService := service.NewAuthService(userRepo, roleRepo, studentRepo, lecturerRepo)
	
	authRoutes := router.Group("/auth")
	
	authRoutes.Post("/login", authService.Login)
	authRoutes.Post("/refresh", authService.RefreshToken)
	authRoutes.Post("/logout", authService.Logout)
	authRoutes.Get("/profile", middleware.RequireAuth(userRepo),authService.Profile,)
}