package routes

import (
	"backend/internal/di"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, container *di.Container) {
	//Логирование запросов
	app.Use(container.RequestMiddleware)

	//Группировка всех api запросов
	api := app.Group("/api")

	//Health check
	api.Get("/", container.HealthHandler.Health)

	// User routes
	userRoutes := api.Group("/user")
	userRoutes.Post("/create", container.UserHandler.CreateUser)
}
