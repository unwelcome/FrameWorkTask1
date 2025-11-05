package main

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	app.Get("/health", func(c *fiber.Ctx) error {
		authServiceHealth := "unknown"

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"gateway": "healthy", "auth": authServiceHealth})
	})

	fmt.Printf("Server listen on http://%s", "localhost:8080")

	if err := app.Listen("localhost:8080"); err != nil {
		fmt.Printf("Failed to start http server: %v", err)
		os.Exit(1)
	}
}
