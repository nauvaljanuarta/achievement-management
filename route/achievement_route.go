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
	// Initialize MongoDB connection
	mongoDB := database.GetMongoDB()
	
	// Create achievement repositories
	achievementRepo := repository.NewAchievementRepository(mongoDB)
	achievementRefRepo := repository.NewAchievementReferenceRepository(database.PgDB)
	
	// Create achievement service
	achievementService := service.NewAchievementService(
		achievementRepo,
		achievementRefRepo,
		studentRepo,
		lecturerRepo,
		userRepo,
		roleRepo, 
	)

	// ===== ACHIEVEMENT ROUTES =====
	achievementRoutes := router.Group("/achievements")
	
	// === LIST ACHIEVEMENTS (Role-based) ===
	// GET /api/v1/achievements // List (filtered by role)
	achievementRoutes.Get("/",
		middleware.RequireAuth(userRepo),
		middleware.RequirePermission("achievement:read"),
		achievementService.GetAchievementsByRole) // Butuh method baru ini
	
	// === DETAIL ACHIEVEMENT ===
	// GET /api/v1/achievements/:id // Detail
	achievementRoutes.Get("/:id",
		middleware.RequireAuth(userRepo),
		middleware.RequirePermission("achievement:read"),
		achievementService.GetAchievementByID)
	
	// === CREATE ACHIEVEMENT ===
	// POST /api/v1/achievements // Create (Mahasiswa)
	achievementRoutes.Post("/",
		middleware.RequireAuth(userRepo),
		middleware.RequireRole("Mahasiswa", roleRepo),
		middleware.RequirePermission("achievement:create"),
		achievementService.CreateAchievement)
	
	// === UPDATE ACHIEVEMENT ===
	// PUT /api/v1/achievements/:id // Update (Mahasiswa)
	achievementRoutes.Put("/:id",
		middleware.RequireAuth(userRepo),
		middleware.RequireRole("Mahasiswa", roleRepo),
		middleware.RequirePermission("achievement:update"),
		achievementService.UpdateAchievement)
	
	// === DELETE ACHIEVEMENT ===
	// DELETE /api/v1/achievements/:id // Delete (Mahasiswa)
	achievementRoutes.Delete("/:id",
		middleware.RequireAuth(userRepo),
		middleware.RequireRole("Mahasiswa", roleRepo),
		middleware.RequirePermission("achievement:delete"),
		achievementService.DeleteAchievement)
	
	// === SUBMIT FOR VERIFICATION ===
	// POST /api/v1/achievements/:id/submit // Submit for verification
	achievementRoutes.Post("/:id/submit",
		middleware.RequireAuth(userRepo),
		middleware.RequireRole("Mahasiswa", roleRepo),
		middleware.RequirePermission("achievement:update"),
		achievementService.SubmitAchievement)
	
	// === VERIFICATION (Dosen Wali) ===
	// POST /api/v1/achievements/:id/verify // Verify (Dosen Wali)
	achievementRoutes.Post("/:id/verify",
		middleware.RequireAuth(userRepo),
		middleware.RequireRole("Dosen Wali", roleRepo),
		middleware.RequirePermission("achievement:verify"),
		achievementService.VerifyAchievement)
	
	// === REJECT (Dosen Wali) ===
	// POST /api/v1/achievements/:id/reject // Reject (Dosen Wali)
	achievementRoutes.Post("/:id/reject",
		middleware.RequireAuth(userRepo),
		middleware.RequireRole("Dosen Wali", roleRepo),
		middleware.RequirePermission("achievement:verify"),
		achievementService.RejectAchievement)
	
	// === STATUS HISTORY ===
	// GET /api/v1/achievements/:id/history // Status history
	achievementRoutes.Get("/:id/history",
		middleware.RequireAuth(userRepo),
		middleware.RequirePermission("achievement:read"),
		achievementService.GetAchievementHistory) // Butuh method baru ini
	
	// === ATTACHMENT UPLOAD ===
	// POST /api/v1/achievements/:id/attachments // Upload files
	achievementRoutes.Post("/:id/attachments",
		middleware.RequireAuth(userRepo),
		middleware.RequirePermission("achievement:update"),
		achievementService.UploadAttachments) // Butuh method baru ini
}