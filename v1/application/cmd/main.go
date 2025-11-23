package main

import (
	"fmt"
	"net"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	application_proto "github.com/unwelcome/FrameWorkTask1/v1/application/api"
	"github.com/unwelcome/FrameWorkTask1/v1/application/internal/config"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/v1/application/internal/logger"
	"github.com/unwelcome/FrameWorkTask1/v1/application/internal/services"
	"google.golang.org/grpc"
)

func main() {
	// Инициализация конфига
	cfg := config.NewConfig()
	cfg.Print()

	// Инициализация логгера
	loggerConf := logger.Setup(cfg.ApplicationService.LogPath, cfg.App.LogConsoleOut)
	log.Logger = *loggerConf

	// Подключение к Postgresql
	db := postgresDB.NewDatabaseInstance(cfg.GetDBConnectionString())

	// Создание сервера
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ApplicationService.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start tcp server")
	}

	// Подключение grpc
	grpcServer := grpc.NewServer()
	application_proto.RegisterApplicationServiceServer(grpcServer, services.NewApplicationService(db))

	// Запуск сервиса
	log.Info().Int("port", cfg.ApplicationService.Port).Msg("Application service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to start grpc server")
	}
}
