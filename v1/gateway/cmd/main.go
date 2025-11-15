package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	_ "github.com/unwelcome/FrameWorkTask1/v1/gateway/api/docs"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/app"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/config"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/logger"
	"github.com/unwelcome/FrameWorkTask1/v1/gateway/internal/routes"
)

// @title     Framework task 2 API
// @version   2.0
// @host      4aik.ru
// @BasePath  /api
// @securityDefinitions.apikey ApiKeyAuth
// @in 		  header
// @name 	  Authorization
func main() {
	// Инициализация конфига
	cfg := config.NewConfig()
	cfg.Print()

	// Инициализация логгера
	loggerConf := logger.Setup(cfg.Gateway.LogPath, cfg.App.LogConsoleOut)
	log.Logger = *loggerConf

	// Инициализация fiber сервера
	server := fiber.New()

	// Инициализация зависимостей
	application := app.InitApp(cfg)

	// Инициализация маршрутов
	routes.SetupRoutes(server, application)

	// Запуск сервера
	log.Info().Msgf("Server listen on http://localhost:%d/api/", cfg.Gateway.Port)
	if err := server.Listen(fmt.Sprintf(":%d", cfg.Gateway.Port)); err != nil {
		log.Fatal().Err(err).Msg("Failed to start http server")
	}
}
