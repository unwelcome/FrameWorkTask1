package di

import (
	"backend/internal/handlers"
	"backend/internal/middlewares"
	"backend/internal/repositories"
	"backend/internal/services"
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type Container struct {
	//Middlewares
	RequestMiddleware func(ctx *fiber.Ctx) error

	//Health
	HealthHandler *handlers.HealthHandler

	//User
	userRepo    *repositories.UserRepository
	userService *services.UserService
	UserHandler *handlers.UserHandler
}

func NewContainer(postgres *sql.DB, logger *slog.Logger) *Container {
	container := &Container{}

	//Инициализация middleware
	container.InitMiddlewares(logger)

	//Инициализация репозиториев
	container.InitRepositories(postgres)

	//Инициализация сервисов
	container.InitServices()

	//Инициализация хендлеров
	container.InitHandlers(logger)

	return container
}

func (c *Container) InitMiddlewares(logger *slog.Logger) {
	c.RequestMiddleware = middlewares.RequestLog(logger)
}

func (c *Container) InitRepositories(postgres *sql.DB) {
	c.userRepo = repositories.NewUserRepository(postgres)
}

func (c *Container) InitServices() {
	c.userService = services.NewUserService(c.userRepo)
}

func (c *Container) InitHandlers(logger *slog.Logger) {
	c.HealthHandler = handlers.NewHealthHandler()
	c.UserHandler = handlers.NewUserHandler(c.userService, logger)
}
