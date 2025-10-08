package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "github.com/unwelcome/iqjtest/api/docs"
	"github.com/unwelcome/iqjtest/database"
	"github.com/unwelcome/iqjtest/internal/config"
	"github.com/unwelcome/iqjtest/internal/dependency_injection"
	"github.com/unwelcome/iqjtest/internal/routes"
)

// @title           IQJ Test Task
// @version         1.0
// @description     Swagger for IQJ test task.

// @host      localhost:8080
// @BasePath  /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Инициализация логгера
	zerolog.TimeFieldFormat = "02.01.2006 15:04:05.000"
	logger := log.With().Logger()

	// Инициализация конфига
	cfg := config.LoadConfig(logger)

	// Подключение к базам данных
	postgres, redis, minio := database.ConnectToDatabases(cfg, logger)
	defer postgres.Close()
	defer redis.Close()

	// Инициализация fiber
	app := fiber.New()

	// Создание контейнера с dependency injection
	container := dependency_injection.NewContainer(postgres, redis, minio, cfg, logger)

	// Инициализация роутов
	routes.SetupRoutes(app, container)

	// Запуск приложения
	if err := app.Listen(cfg.GetAppInternalAddress()); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}
