package main

import (
	"fmt"
	"net"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
	company_proto "github.com/unwelcome/FrameWorkTask1/v1/company/api"
	"github.com/unwelcome/FrameWorkTask1/v1/company/internal/config"
	postgresDB "github.com/unwelcome/FrameWorkTask1/v1/company/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/v1/company/internal/logger"
	"github.com/unwelcome/FrameWorkTask1/v1/company/internal/services"
	"google.golang.org/grpc"
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
	// cache := redisDB.NewCacheInstance(cfg.GetCacheConnectionOptions(), cfg.App.RefreshTokenLifetime)

	// Создание сервера
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.CompanyService.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start tcp server")
	}

	// Подключение grpc
	grpcServer := grpc.NewServer()
	company_proto.RegisterCompanyServiceServer(grpcServer, services.NewCompanyService(db))

	// Запуск сервиса
	log.Info().Int("port", cfg.CompanyService.Port).Msg("Company service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to start grpc server")
	}
}
