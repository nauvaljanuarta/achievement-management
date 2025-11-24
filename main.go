package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"achievement-backend/config"   
	"achievement-backend/database"
	// "achievement-backend/route"
)

// @title Student Achievement System API
// @version 1.0
// @description API Sistem Pelaporan Prestasi Mahasiswa
// @host localhost:3000
// @BasePath /api/v1
func main() {
	if err := config.LoadConfig(); err != nil {
		log.Println("Warning: .env file not found")
	}

	database.ConnectDB()
	defer database.CloseDB()

	app := fiber.New(config.FiberConfig())

	app.Use(recover.New())
	app.Use(cors.New())
	app.Use(logger.New(config.LoggerConfig()))

	// 5. Setup Routes
	// route.SetupRoutes(app)

	port := config.GetEnv("APP_PORT", "3000")
	
	log.Printf("Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}