package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/application/api/generated"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/auth/api/generated"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/company/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/config"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/handlers"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/middlewares"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	OperationIDKey = "operation_id"
	UserUUIDKey    = "user_uuid"
)

type App struct {
	AuthServiceClient        auth_proto.AuthServiceClient
	CompanyServiceClient     company_proto.CompanyServiceClient
	ApplicationServiceClient application_proto.ApplicationServiceClient

	OperationIDMiddleware fiber.Handler
	LoggerMiddleware      fiber.Handler
	AuthMiddleware        fiber.Handler

	HealthHandler      handlers.HealthHandler
	AuthHandler        handlers.AuthHandler
	CompanyHandler     handlers.CompanyHandler
	ApplicationHandler handlers.ApplicationHandler
}

func InitApp(cfg *config.Config) *App {
	application := &App{}

	application.AuthServiceClient = auth_proto.NewAuthServiceClient(dial(cfg.Auth.Addr()))
	application.CompanyServiceClient = company_proto.NewCompanyServiceClient(dial(cfg.Company.Addr()))
	application.ApplicationServiceClient = application_proto.NewApplicationServiceClient(dial(cfg.App.Addr()))

	application.OperationIDMiddleware = middlewares.NewOperationIDMiddleware(OperationIDKey)
	application.LoggerMiddleware = middlewares.NewRequestLoggerMiddleware(OperationIDKey)
	application.AuthMiddleware = middlewares.NewAuthMiddleware(cfg.JWT.Secret, UserUUIDKey)

	application.HealthHandler = handlers.NewHealthHandler(application.AuthServiceClient, application.CompanyServiceClient, application.ApplicationServiceClient, OperationIDKey)
	application.AuthHandler = handlers.NewAuthHandler(application.AuthServiceClient, OperationIDKey, UserUUIDKey)
	application.CompanyHandler = handlers.NewCompanyHandler(application.CompanyServiceClient, OperationIDKey, UserUUIDKey)
	application.ApplicationHandler = handlers.NewApplicationHandler(application.ApplicationServiceClient, OperationIDKey, UserUUIDKey)

	return application
}

func dial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Str("addr", addr).Msg("failed to connect to service")
	}
	return conn
}
