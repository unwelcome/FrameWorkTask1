package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"time"
)

func main() {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		fmt.Println(c.Method(), "request:", c.IP(), "at", time.Now().Format("2006-01-02 15:04:05"))
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, Go!")
	})
	app.Get("/download", func(c *fiber.Ctx) error {
		return c.Download("./main.go")
	})

	app.Listen(":8080")

	// test
}
