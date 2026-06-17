package app

import (
	fiberprometheus "github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
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

	PrometheusMiddleware  fiber.Handler
	OperationIDMiddleware fiber.Handler
	LoggerMiddleware      fiber.Handler
	AuthMiddleware        fiber.Handler

	PasswordRateLimiter fiber.Handler
	CodeRateLimiter     fiber.Handler
	UserRateLimiter     fiber.Handler

	HealthHandler      handlers.HealthHandler
	AuthHandler        handlers.AuthHandler
	CompanyHandler     handlers.CompanyHandler
	ApplicationHandler handlers.ApplicationHandler
}

func InitApp(cfg *config.Config, httpLogger zerolog.Logger, redisClient *redis.Client) *App {
	// Загружаем публичный ключ из PEM-файла
	publicKey, err := utils.LoadPublicKey(cfg.JWT.PublicKeyPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", cfg.JWT.PublicKeyPath).Msg("failed to load JWT public key")
	}

	application := &App{AppEnv: cfg.AppEnv}

	// Подключение к сервисам
	application.AuthServiceClient = auth_proto.NewAuthServiceClient(dial(cfg.Auth.Addr()))
	application.CompanyServiceClient = company_proto.NewCompanyServiceClient(dial(cfg.Company.Addr()))
	application.ApplicationServiceClient = application_proto.NewApplicationServiceClient(dial(cfg.App.Addr()))

	// Инициализация prometheus middleware
	fp := fiberprometheus.New("gateway")
	application.PrometheusMiddleware = fp.Middleware

	// Инициализация остальных middleware
	application.OperationIDMiddleware = middlewares.NewOperationIDMiddleware(OperationIDKey)
	application.LoggerMiddleware = middlewares.NewRequestLoggerMiddleware(OperationIDKey, UserUUIDKey, httpLogger)
	application.AuthMiddleware = middlewares.NewAuthMiddleware(publicKey, UserUUIDKey)

	// Инициализация rate limiter-ов
	ipKey := func(c *fiber.Ctx) string { return c.IP() }
	userKey := func(c *fiber.Ctx) string { return utils.GetLocal[string](c, UserUUIDKey) }

	application.PasswordRateLimiter = middlewares.NewRateLimiter(redisClient, cfg.Redis.Prefix, "password", middlewares.RateLimiterConfig{
		Capacity:     cfg.RateLimit.Password.Capacity,
		RefillPerSec: cfg.RateLimit.Password.RefillPerSec,
		KeyFn:        ipKey,
	})
	application.CodeRateLimiter = middlewares.NewRateLimiter(redisClient, cfg.Redis.Prefix, "code", middlewares.RateLimiterConfig{
		Capacity:     cfg.RateLimit.Code.Capacity,
		RefillPerSec: cfg.RateLimit.Code.RefillPerSec,
		KeyFn:        ipKey,
	})
	application.UserRateLimiter = middlewares.NewRateLimiter(redisClient, cfg.Redis.Prefix, "user", middlewares.RateLimiterConfig{
		Capacity:     cfg.RateLimit.User.Capacity,
		RefillPerSec: cfg.RateLimit.User.RefillPerSec,
		KeyFn:        userKey,
	})

	// Инициализация Geo
	sessionProvider := session.New(cfg.GeoIP.CityDBPath, cfg.GeoIP.ASNDBPath, log.Logger)

	// Инициализация handler-ов
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
