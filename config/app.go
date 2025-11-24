package config

import (
	"time"
	"github.com/gofiber/fiber/v2"
)

func FiberConfig() fiber.Config {
	return fiber.Config{
		AppName:      "Student Achievement System v1",
		ReadTimeout:  10 * time.Second, // handle timeouts
		WriteTimeout: 10 * time.Second,
		BodyLimit:    10 * 1024 * 1024, // limit body size to 10MB
		JSONEncoder:  nil,              
		JSONDecoder:  nil,
	}
}