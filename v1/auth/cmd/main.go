package main

import (
	"fmt"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/api"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/internal/config"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/auth/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/internal/logger"
	"github.com/unwelcome/FrameWorkTask1/v1/auth/internal/services"
	"google.golang.org/grpc"
	"net"
)

func main() {
	// Инициализация конфига
	cfg := config.NewConfig()
	cfg.Print()

	// Инициализация логгера
	loggerConf := logger.Setup(cfg.App.LogPath, cfg.App.LogConsoleOut)
	log.Logger = *loggerConf

	// Подключение к Postgresql
	db := postgresDB.NewDatabaseInstance(cfg.GetDBConnectionString())
	// Подключение к Redis

	// Создание сервера
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.AuthService.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start tcp server")
	}

	// Подключение grpc
	grpcServer := grpc.NewServer()
	auth_proto.RegisterAuthServiceServer(grpcServer, services.NewAuthService(db))

	// Запуск сервиса
	log.Info().Int("port", cfg.AuthService.Port).Msg("Auth service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to start grpc server")
	}
}
