package route

import (
	"achievement-backend/middleware"
	"achievement-backend/app/repository"
	"achievement-backend/app/service"

	"github.com/gofiber/fiber/v2"
)

func setupStudentLecturerRoutes(
	router fiber.Router,
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	achievementRepo repository.AchievementRepository,
	achievementRefRepo repository.AchievementReferenceRepository,
	roleRepo repository.RoleRepository,

) {
	studentLecturerService := service.NewStudentLecturerService(
		studentRepo,
		lecturerRepo,
		userRepo,
		roleRepo,
		achievementRepo,
		achievementRefRepo,
	)

	studentRoutes := router.Group("/students")
	studentProtected := studentRoutes.Group("", middleware.RequireAuth(userRepo))
	
	studentProtected.Get("/", studentLecturerService.GetAllStudents)
	studentProtected.Get("/:id", studentLecturerService.GetStudentByID)
	studentProtected.Get("/:id/achievements", middleware.RequirePermission("achievement:read"), studentLecturerService.GetStudentAchievements)
	
	// PUT /students/:id/advisor - hanya Admin
	studentProtected.Put("/:id/advisor", studentLecturerService.UpdateStudentAdvisor)

	lecturerRoutes := router.Group("/lecturers")
	lecturerProtected := lecturerRoutes.Group("", middleware.RequireAuth(userRepo))
	lecturerProtected.Get("/", studentLecturerService.GetAllLecturers)
	// GET /lecturers/:id/advisees - Admin atau Dosen Wali itu sendiri
	lecturerProtected.Get("/:id/advisees",  studentLecturerService.GetLecturerAdvisees)
}