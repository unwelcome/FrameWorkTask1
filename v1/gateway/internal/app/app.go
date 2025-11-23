package app

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	application_proto "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/application"
	auth_proto "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/auth"
	company_proto "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/company"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/config"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/handlers"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/middlewares"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	OperationIDKey = "operation_id"
	UserUUIDKey    = "user_uuid"
)

type App struct {
	// Services
	AuthServiceClient        auth_proto.AuthServiceClient
	CompanyServiceClient     company_proto.CompanyServiceClient
	ApplicationServiceClient application_proto.ApplicationServiceClient

	// Middlewares
	OperationIDMiddleware fiber.Handler
	LoggerMiddleware      fiber.Handler
	AuthMiddleware        fiber.Handler

	// Handlers
	HealthHandler  handlers.HealthHandler
	AuthHandler    handlers.AuthHandler
	CompanyHandler handlers.CompanyHandler
}

func InitApp(cfg *config.Config) *App {
	app := &App{}

	// Init clients to services
	app.AuthServiceClient = auth_proto.NewAuthServiceClient(getGRPCConnection(cfg.AuthService.Host, cfg.AuthService.Port))
	app.CompanyServiceClient = company_proto.NewCompanyServiceClient(getGRPCConnection(cfg.CompanyService.Host, cfg.CompanyService.Port))
	app.ApplicationServiceClient = application_proto.NewApplicationServiceClient(getGRPCConnection(cfg.ApplicationService.Host, cfg.ApplicationService.Port))

	// Init middlewares
	app.OperationIDMiddleware = middlewares.NewOperationIDMiddleware(OperationIDKey)
	app.LoggerMiddleware = middlewares.NewRequestLoggerMiddleware(OperationIDKey)
	app.AuthMiddleware = middlewares.NewAuthMiddleware(cfg.App.JWTSecret, UserUUIDKey)

	// Init handlers
	app.HealthHandler = handlers.NewHealthHandler(app.AuthServiceClient, app.CompanyServiceClient, app.ApplicationServiceClient, OperationIDKey)
	app.AuthHandler = handlers.NewAuthHandler(app.AuthServiceClient, OperationIDKey, UserUUIDKey)
	app.CompanyHandler = handlers.NewCompanyHandler(app.CompanyServiceClient, OperationIDKey, UserUUIDKey)

	return app
}

func getGRPCConnection(host string, port int) *grpc.ClientConn {
	endpoint := fmt.Sprintf("%s:%d", host, port)

	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to connect to %s:%d", host, port)
	}
	//defer conn.Close()

	return conn
}
