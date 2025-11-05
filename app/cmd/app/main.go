package main

import (
	"microapi/internal/app"
	"microapi/internal/config"
	"microapi/internal/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Init logger
	log := logger.MustLogger()
	log.Info().Msg("Logger initialized")

	// Init config
	cfg := config.MustConfig(log)
	log.Info().Msg("Config initialized")

	// Init application
	application := app.NewApp(log, cfg)
	log.Info().Msg("Application initialized")

	// Start gRPC services async
	go application.GRPCServer.MustRun()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	stopSign := <-stop

	log.Info().Str("signal", stopSign.String()).Msg("Received stop signal")

	application.GRPCServer.Stop()

	log.Info().Msg("Application stopped")
}
