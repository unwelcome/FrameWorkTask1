package dependency_injection

import (
	"database/sql"
	"github.com/minio/minio-go/v7"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/unwelcome/iqjtest/internal/config"
	"github.com/unwelcome/iqjtest/internal/handlers"
	"github.com/unwelcome/iqjtest/internal/middlewares"
	"github.com/unwelcome/iqjtest/internal/repositories"
	"github.com/unwelcome/iqjtest/internal/services"
)

type Container struct {
	// Middleware
	LoggingMiddleware      func(c *fiber.Ctx) error
	AuthMiddleware         func(c *fiber.Ctx) error
	CatOwnershipMiddleware func(c *fiber.Ctx) error

	// Health
	HealthHandler handlers.HealthHandler

	// Auth
	authRepository repositories.AuthRepository
	authService    services.AuthService
	AuthHandler    handlers.AuthHandler

	// User
	userRepository repositories.UserRepository
	userService    services.UserService
	UserHandler    handlers.UserHandler

	// Cat
	catRepository repositories.CatRepository
	catService    services.CatService
	CatHandler    handlers.CatHandler

	// CatPhoto
	catPhotoRepository repositories.CatPhotoRepository
	catPhotoService    services.CatPhotoService
	CatPhotoHandler    handlers.CatPhotoHandler
}

func NewContainer(postgres *sql.DB, redis *redis.Client, minio *minio.Client, cfg *config.Config, logger zerolog.Logger) *Container {
	// Создание контейнера
	container := &Container{}

	// Инициализация репозиториев
	container.InitRepositories(postgres, redis, minio, cfg)

	// Инициализация сервисов
	container.InitServices(cfg)

	// Инициализация хендлеров
	container.InitHandlers(cfg)

	// Инициализация middleware
	container.InitMiddlewares(logger, cfg) // Инициализируем после инициализации сервисов

	return container
}

func (c *Container) InitMiddlewares(logger zerolog.Logger, cfg *config.Config) {
	c.LoggingMiddleware = middlewares.LoggingRequest(logger)
	c.AuthMiddleware = middlewares.AuthMiddleware(cfg.JWTSecret)
	c.CatOwnershipMiddleware = middlewares.CatOwnershipMiddleware(c.catService, cfg.Timeouts.Middleware)
}

func (c *Container) InitRepositories(postgres *sql.DB, redis *redis.Client, minio *minio.Client, cfg *config.Config) {
	c.userRepository = repositories.NewUserRepository(postgres)
	c.authRepository = repositories.NewAuthRepository(redis)
	c.catRepository = repositories.NewCatRepository(postgres)
	c.catPhotoRepository = repositories.NewCatPhotoRepository(postgres, minio, cfg.S3ConnConfig().PublicEndpoint, cfg.S3Buckets["catPhotoBucket"].Name)
}

func (c *Container) InitServices(cfg *config.Config) {
	c.userService = services.NewUserService(c.userRepository, cfg.BCryptCost)
	c.authService = services.NewAuthService(c.userService, c.authRepository, cfg.JWTSecret, cfg.AccessTokenLifetime, cfg.RefreshTokenLifetime)
	c.catPhotoService = services.NewCatPhotoService(c.catPhotoRepository)
	c.catService = services.NewCatService(c.catRepository, c.catPhotoService)
}

func (c *Container) InitHandlers(cfg *config.Config) {
	c.HealthHandler = handlers.NewHealthHandler()
	c.UserHandler = handlers.NewUserHandler(c.userService, cfg.Timeouts.Request)
	c.AuthHandler = handlers.NewAuthHandler(c.authService, cfg.Timeouts.Request)
	c.CatHandler = handlers.NewCatHandler(c.catService, cfg.Timeouts.Request, cfg.Timeouts.FileRequest)
	c.CatPhotoHandler = handlers.NewCatPhotoHandler(c.catPhotoService, cfg.Timeouts.Request, cfg.Timeouts.FileRequest)
}
