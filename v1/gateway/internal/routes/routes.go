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
	auth := api.Group("/auth")

	// Health handler
	api.Get("/health", app.HealthHandler.Health)

	// Auth handler
	api.Post("/register", app.AuthHandler.Register)
	api.Post("/login", app.AuthHandler.Login)
	api.Post("/refresh", app.AuthHandler.RefreshToken)
	auth.Get("/user/:user_uuid/info", app.AuthHandler.GetUser)
	auth.Get("/user/tokens", app.AuthHandler.GetAllActiveTokens)
	auth.Patch("/user/password", app.AuthHandler.ChangePassword)
	auth.Patch("/user/bio", app.AuthHandler.UpdateUserBio)
	auth.Delete("/user/account", app.AuthHandler.DeleteUser)
	auth.Delete("/user/revoke/token", app.AuthHandler.RevokeToken)
	auth.Delete("/user/revoke/all", app.AuthHandler.RevokeAllTokens)

	// Company handler
	auth.Post("/company/create", app.CompanyHandler.CreateCompany)
}
