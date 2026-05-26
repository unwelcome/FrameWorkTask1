package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
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
	cfg := config.NewConfig()

	loggerConf, httpLogger := logger.Setup(cfg.Log.Path, cfg.Log.ConsoleOut)
	log.Logger = *loggerConf

	server := fiber.New()

	application := app.InitApp(cfg, *httpLogger)

	routes.SetupRoutes(server, application)

	log.Info().Msgf("server listening on http://localhost:%d/api/", cfg.Port)
	if err := server.Listen(fmt.Sprintf(":%d", cfg.Port)); err != nil {
		log.Fatal().Err(err).Msg("failed to start http server")
	}
}
