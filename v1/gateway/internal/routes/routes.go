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
	api.Post("/login", app.AuthHandler.Login)
	api.Post("/refresh", app.AuthHandler.RefreshToken)
	auth.Get("/user/tokens", app.AuthHandler.GetAllActiveTokens)
	auth.Delete("/user/revoke/token", app.AuthHandler.RevokeToken)
	auth.Delete("/user/revoke/all", app.AuthHandler.RevokeAllTokens)
	// User
	api.Post("/register", app.AuthHandler.Register)
	auth.Get("/user/:user_uuid/info", app.AuthHandler.GetUser)
	auth.Patch("/user/password", app.AuthHandler.ChangePassword)
	auth.Patch("/user/bio", app.AuthHandler.UpdateUserBio)
	auth.Delete("/user/account", app.AuthHandler.DeleteUser)

	// Company handler
	auth.Post("/company/create", app.CompanyHandler.CreateCompany)
	auth.Get("/company/:company_uuid", app.CompanyHandler.GetCompany)
	auth.Get("/company/list", app.CompanyHandler.GetCompanies)
	auth.Patch("/company/:company_uuid/title", app.CompanyHandler.UpdateCompanyTitle)
	auth.Patch("/company/:company_uuid/status", app.CompanyHandler.UpdateCompanyStatus)
	auth.Delete("/company/:company_uuid", app.CompanyHandler.DeleteCompany)
	// Join code
	auth.Post("/company/:company_uuid/code", app.CompanyHandler.CreateCompanyJoinCode)
	auth.Get("/company/:company_uuid/codes", app.CompanyHandler.GetCompanyJoinCodes)
	auth.Delete("/company/:company_uuid/code", app.CompanyHandler.DeleteCompanyJoinCode)
	// Employee
	auth.Post("/company/join", app.CompanyHandler.JoinCompany)
	auth.Get("/company/:company_uuid/employee/:employee_uuid/info", app.CompanyHandler.GetCompanyEmployee)
	auth.Get("/company/:company_uuid/employees/summary", app.CompanyHandler.GetCompanyEmployeesSummary)
	auth.Delete("/company/:company_uuid/employee/:employee_uuid", app.CompanyHandler.RemoveCompanyEmployee)
}
