package routes

import (
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/app"
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
	api.Post("/user/verify", app.AuthHandler.VerifyAccount)
	api.Post("/user/verify/resend", app.AuthHandler.ResendVerificationCode)
	api.Post("/forgot-password", app.AuthHandler.ForgotPassword)
	api.Post("/reset-password", app.AuthHandler.ResetPassword)
	api.Post("/verify-2fa", app.AuthHandler.Verify2FA)
	// Debug-only routes — доступны только при APP_ENV=test
	if app.AppEnv == "test" {
		api.Get("/debug/user/:user_uuid/verification-code", app.AuthHandler.GetVerificationCode)
		api.Get("/debug/user/:user_uuid/recovery-code", app.AuthHandler.GetRecoveryCode)
		api.Get("/debug/2fa/:session_uuid/code", app.AuthHandler.Get2FACode)
	}
	// Tokens
	auth.Get("/user/tokens", app.AuthHandler.GetAllActiveTokens)
	auth.Delete("/user/revoke/token", app.AuthHandler.RevokeToken)
	auth.Delete("/user/revoke/all", app.AuthHandler.RevokeAllTokens)
	// User
	api.Post("/register", app.AuthHandler.Register)
	api.Post("/restore-account", app.AuthHandler.RestoreAccount)
	auth.Get("/user/:user_uuid/info", app.AuthHandler.GetUser)
	auth.Patch("/user/password", app.AuthHandler.ChangePassword)
	auth.Patch("/user/bio", app.AuthHandler.UpdateUserBio)
	auth.Patch("/user/2fa", app.AuthHandler.UpdateUser2FA)
	auth.Delete("/user/account", app.AuthHandler.DeleteUser)

	// Company handler
	auth.Get("/company/my", app.CompanyHandler.GetUserCompanies)
	auth.Get("/company/list", app.CompanyHandler.GetCompanies)
	auth.Get("/company/:company_uuid", app.CompanyHandler.GetCompany)
	auth.Get("/company/:company_uuid/codes", app.CompanyHandler.GetCompanyJoinCodes)
	auth.Get("/company/:company_uuid/employee/:employee_uuid/info", app.CompanyHandler.GetCompanyEmployee)
	auth.Get("/company/:company_uuid/employees/list", app.CompanyHandler.GetCompanyEmployees)
	auth.Get("/company/:company_uuid/employees/summary", app.CompanyHandler.GetCompanyEmployeesSummary)
	auth.Get("/company/:company_uuid/departments/list", app.CompanyHandler.GetCompanyDepartments)
	auth.Get("/company/:company_uuid/department/:department_uuid", app.CompanyHandler.GetDepartment)
	auth.Post("/company/create", app.CompanyHandler.CreateCompany)
	auth.Post("/company/join", app.CompanyHandler.JoinCompany)
	auth.Post("/company/:company_uuid/code", app.CompanyHandler.CreateCompanyJoinCode)
	auth.Post("/company/:company_uuid/department", app.CompanyHandler.CreateDepartment)
	auth.Post("/company/:company_uuid/department/:department_uuid/employee/:employee_uuid", app.CompanyHandler.AddEmployeeToDepartment)
	auth.Patch("/company/:company_uuid/title", app.CompanyHandler.UpdateCompanyTitle)
	auth.Patch("/company/:company_uuid/status", app.CompanyHandler.UpdateCompanyStatus)
	auth.Patch("/company/:company_uuid/employee/:employee_uuid/role", app.CompanyHandler.UpdateEmployeeRole)
	auth.Patch("/company/:company_uuid/department/:department_uuid/title", app.CompanyHandler.UpdateDepartmentTitle)
	auth.Delete("/company/:company_uuid", app.CompanyHandler.DeleteCompany)
	auth.Delete("/company/:company_uuid/code", app.CompanyHandler.DeleteCompanyJoinCode)
	auth.Delete("/company/:company_uuid/employee/:employee_uuid", app.CompanyHandler.RemoveCompanyEmployee)
	auth.Delete("/company/:company_uuid/department/:department_uuid", app.CompanyHandler.DeleteDepartment)
	auth.Delete("/company/:company_uuid/department/:department_uuid/employee/:employee_uuid", app.CompanyHandler.RemoveEmployeeFromDepartment)

	// Application handler
	auth.Get("/application/:application_uuid", app.ApplicationHandler.GetApplication)
	auth.Get("/company/:company_uuid/applications/list", app.ApplicationHandler.GetApplications)
	auth.Post("/application/create", app.ApplicationHandler.CreateApplication)
	auth.Post("/application/:application_uuid/fix-log", app.ApplicationHandler.AddApplicationFixLog)
	auth.Patch("/application/:application_uuid/status", app.ApplicationHandler.UpdateApplicationStatus)
	auth.Patch("/application/:application_uuid/assign", app.ApplicationHandler.AssignApplication)
	auth.Patch("/application/:application_uuid/redirect", app.ApplicationHandler.RedirectApplication)
	auth.Patch("/application/:application_uuid/recall", app.ApplicationHandler.RecallApplication)
	auth.Patch("/application/:application_uuid/take-verification", app.ApplicationHandler.TakeApplicationToVerification)
	auth.Patch("/application/:application_uuid/release-verification", app.ApplicationHandler.ReleaseApplicationVerification)
	auth.Delete("/application/:application_uuid", app.ApplicationHandler.DeleteApplication)
	auth.Get("/application/:application_uuid/history", app.ApplicationHandler.GetApplicationHistory)
}
