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