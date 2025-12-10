package route

import (
    "achievement-backend/database"
    "achievement-backend/app/repository"
    "achievement-backend/app/service"

    "github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
    db := database.PgDB
    
    userRepo := repository.NewUserRepository(db)
    roleRepo := repository.NewRoleRepository(db)
    studentRepo := repository.NewStudentRepository(db)
    lecturerRepo := repository.NewLecturerRepository(db)
    
    userService := service.NewUserService(userRepo, roleRepo, studentRepo, lecturerRepo)
    
    examAPI := app.Group("/exam/api")
    
    setupAuthRoutes(examAPI, userRepo, roleRepo)
    setupUserRoutes(examAPI, userService, userRepo, roleRepo)
    
    examAPI.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status":  "OK",
            "service": "Student Achievement System API",
            "version": "1.0",
        })
    })
}