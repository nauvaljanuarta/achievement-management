package route

import (
	"achievement-backend/middleware"
	"achievement-backend/app/repository"
	"achievement-backend/app/service"
	"achievement-backend/database"

	"github.com/gofiber/fiber/v2"
)

func setupAchievementRoutes(
	router fiber.Router,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
) {
	mongoDB := database.GetMongoDB()
	
	achievementRepo := repository.NewAchievementRepository(mongoDB)
	achievementRefRepo := repository.NewAchievementReferenceRepository(database.PgDB)
	
	achievementService := service.NewAchievementService(
		achievementRepo,
		achievementRefRepo,
		studentRepo,
		lecturerRepo,
		userRepo,
		roleRepo,
	)

	achievementRoutes := router.Group("/achievements")
	
	protectedRoutes := achievementRoutes.Group("", middleware.RequireAuth(userRepo))
	
	protectedRoutes.Get("/", middleware.RequirePermission("achievement:read"), achievementService.GetAchievementsByRole) // Filter berdasarkan role di service
	protectedRoutes.Get("/:id", middleware.RequirePermission("achievement:read"), achievementService.GetAchievementByID) // Validasi ownership di service
	protectedRoutes.Get("/:id/history", middleware.RequirePermission("achievement:read"), achievementService.GetAchievementHistory)
	protectedRoutes.Post("/", middleware.RequirePermission("achievement:create"), achievementService.CreateAchievement)
	protectedRoutes.Put("/:id", middleware.RequirePermission("achievement:update"), achievementService.UpdateAchievement) // Validasi draft status & ownership di service
	protectedRoutes.Delete("/:id", middleware.RequirePermission("achievement:delete"), achievementService.DeleteAchievement) // Validasi draft status di service
	protectedRoutes.Post("/:id/submit", middleware.RequirePermission("achievement:update"), achievementService.SubmitAchievement) // Validasi status draft → submitted
	
	protectedRoutes.Post("/:id/verify", middleware.RequirePermission("achievement:verify"), achievementService.VerifyAchievement) // Validasi dosen punya mahasiswa bimbingan
	protectedRoutes.Post("/:id/reject", middleware.RequirePermission("achievement:verify"), achievementService.RejectAchievement)

	protectedRoutes.Post("/:id/attachments", middleware.RequirePermission("achievement:update"),achievementService.UploadAttachments)
	
	mahasiswaOnlyRoutes := protectedRoutes.Group("", 
		middleware.RequireRole("Mahasiswa", roleRepo))
	mahasiswaOnlyRoutes.Get("/my", achievementService.GetMyAchievements)
	
	// HANYA ADMIN - Operations khusus admin
	// adminOnlyRoutes := protectedRoutes.Group("",
	// 	middleware.RequireRole("Admin", roleRepo))
}