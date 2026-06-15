package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	_ "github.com/unwelcome/FrameWorkTask1/backend/gateway/api/docs"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/app"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/config"
	"github.com/unwelcome/FrameWorkTask1/backend/gateway/internal/routes"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/logger"
)

// @title     Framework task 2 API
// @version   1.0
// @host      localhost:8080
// @BasePath  /api
// @securityDefinitions.apikey ApiKeyAuth
// @in 		  header
// @name 	  Authorization
func main() {
	// Инициализация конфига
	cfg := config.NewConfig()

	// Инициализация logger-ов
	loggerConf, httpLogger := logger.Setup(cfg.Log.Path, cfg.Log.ConsoleOut)
	log.Logger = *loggerConf

	// Подключение Redis для rate-limit-ов
	redisClient := redis.NewClient(cfg.Redis.Options())

	// Инициализация fiber-сервера
	server := fiber.New()

	// Инициализация всех зависимостей
	application := app.InitApp(cfg, *httpLogger, redisClient)

	// Настройка маршрутизации
	routes.SetupRoutes(server, application)

	// Запуск сервера
	log.Info().Msgf("server listening on http://localhost:%d/api/", cfg.Port)
	if err := server.Listen(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal().Err(err).Msg("failed to start http server")
	}
}
