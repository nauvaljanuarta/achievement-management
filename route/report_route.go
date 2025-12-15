// route/report_route.go
package route

import (
	"achievement-backend/middleware"
	"achievement-backend/app/repository"
	"achievement-backend/app/service"

	"github.com/gofiber/fiber/v2"
)

func SetupReportRoutes(
	router fiber.Router,
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	lecturerRepo repository.LecturerRepository,
	reportRepo repository.ReportRepository,
	roleRepo repository.RoleRepository,
) {
	reportService := service.NewReportService(
		reportRepo,
		userRepo,
		studentRepo,
		lecturerRepo,
		roleRepo,
	)
	router.Get("/reports/statistics", middleware.RequireAuth(userRepo), reportService.GetStatistics,)
	router.Get("/reports/student/:id", middleware.RequireAuth(userRepo),reportService.GetStudentReport,)
}