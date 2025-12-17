package main

import (
	"flag"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"achievement-backend/config"
	"achievement-backend/database"
	"achievement-backend/route"
)

// @title Achievement Management Backend API
// @version 1.0
// @description Backend API untuk manajemen prestasi mahasiswa
// @termsOfService http://localhost:3000/terms

// @contact.name API Support
// @contact.url http://localhost:3000/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:3000
// @BasePath exam/api/
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	if err := config.LoadConfig(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Flag migrate
	migrateFlag := flag.Bool("migrate", false, "Run database migrations")
	flag.Parse()

	database.ConnectDB()
	defer database.CloseDB()

	if *migrateFlag {
		log.Println("Running migrations...")
		database.Migrate(database.PgDB, "./database/migrations")
		log.Println("Migrations completed")
		return
	}

	app := fiber.New(config.FiberConfig())
	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(logger.New(config.LoggerConfig()))

	// Setup routes
	route.SetupRoutes(app) 

	port := config.GetEnv("APP_PORT", "3000")
	log.Printf("🚀 Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}