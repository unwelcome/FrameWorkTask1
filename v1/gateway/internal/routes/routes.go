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

	// Middleware индексирования операций
	api.Use(app.OperationIDMiddleware)

	// Middleware логирования
	api.Use(app.LoggerMiddleware)

	// Middleware авторизации
	api.Use("/auth", app.AuthMiddleware)

	// Health handler
	api.Get("/health", app.HealthHandler.Health)

	// Auth handler
	api.Post("/register", app.AuthHandler.Register)
	api.Post("/login", app.AuthHandler.Login)
	api.Post("/refresh", app.AuthHandler.RefreshToken)
	api.Get("/auth/user/:user_uuid/info", app.AuthHandler.GetUser)
	api.Get("/auth/user/tokens", app.AuthHandler.GetAllActiveTokens)
	api.Patch("/auth/user/password", app.AuthHandler.ChangePassword)
	api.Patch("/auth/user/bio", app.AuthHandler.UpdateUserBio)
	api.Delete("/auth/user/account", app.AuthHandler.DeleteUser)
	api.Delete("/auth/user/revoke/token", app.AuthHandler.RevokeToken)
	api.Delete("/auth/user/revoke/all", app.AuthHandler.RevokeAllTokens)
}
