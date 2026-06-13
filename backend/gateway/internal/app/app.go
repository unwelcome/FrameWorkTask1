package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/application/generated"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/config"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/handlers"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/middlewares"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/session"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	OperationIDKey = "operation_id"
	UserUUIDKey    = "user_uuid"
)

type App struct {
	AppEnv string

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

func InitApp(cfg *config.Config, httpLogger zerolog.Logger) *App {
	// Загружаем публичный ключ из PEM-файла
	publicKey, err := utils.LoadPublicKey(cfg.JWT.PublicKeyPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", cfg.JWT.PublicKeyPath).Msg("failed to load JWT public key")
	}

	application := &App{AppEnv: cfg.AppEnv}

	application.AuthServiceClient = auth_proto.NewAuthServiceClient(dial(cfg.Auth.Addr()))
	application.CompanyServiceClient = company_proto.NewCompanyServiceClient(dial(cfg.Company.Addr()))
	application.ApplicationServiceClient = application_proto.NewApplicationServiceClient(dial(cfg.App.Addr()))

	application.OperationIDMiddleware = middlewares.NewOperationIDMiddleware(OperationIDKey)
	application.LoggerMiddleware = middlewares.NewRequestLoggerMiddleware(OperationIDKey, UserUUIDKey, httpLogger)
	application.AuthMiddleware = middlewares.NewAuthMiddleware(publicKey, UserUUIDKey)

	sessionProvider := session.New(cfg.GeoIP.CityDBPath, cfg.GeoIP.ASNDBPath, log.Logger)

	application.HealthHandler = handlers.NewHealthHandler(application.AuthServiceClient, application.CompanyServiceClient, application.ApplicationServiceClient, OperationIDKey)
	application.AuthHandler = handlers.NewAuthHandler(application.AuthServiceClient, application.CompanyServiceClient, OperationIDKey, UserUUIDKey, sessionProvider)
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
