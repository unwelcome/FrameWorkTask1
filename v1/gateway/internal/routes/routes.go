package routes

import (
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/app"
)

func SetupRoutes(router *fiber.App, app *app.App) {
	api := router.Group("/api")

	// Инициализация swagger
	// swag init -o ./api/docs --dir ./cmd,./internal/entities,./internal/errors,./internal/handlers
	api.Get("/swagger/*", swagger.HandlerDefault)

	// Middleware логирования
	api.Use(app.LoggerMiddleware)

	// Health handler
	api.Get("/health", app.HealthHandler.Health)
}
