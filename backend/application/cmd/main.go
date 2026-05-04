package main

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/application/api/generated"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/services"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/company/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/config"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Инициализация конфига
	cfg := config.NewConfig()
	cfg.Print()

	// Инициализация логгера
	loggerConf := logger.Setup(cfg.ApplicationService.LogPath, cfg.App.LogConsoleOut)
	log.Logger = *loggerConf

	// Подключение к Postgresql
	db := postgresDB.NewDatabaseInstance(cfg.GetDBConnectionString(cfg.ApplicationService.ServiceConfig))

	// Подключение к company сервису
	companyAddr := fmt.Sprintf("%s:%d", cfg.CompanyService.Host, cfg.CompanyService.Port)
	companyConn, err := grpc.NewClient(companyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Str("addr", companyAddr).Msg("Failed to connect to company service")
	}
	defer companyConn.Close()

	companyClient := company_proto.NewCompanyServiceClient(companyConn)

	// Создание сервера
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ApplicationService.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start tcp server")
	}

	// Подключение grpc
	grpcServer := grpc.NewServer()
	application_proto.RegisterApplicationServiceServer(grpcServer, services.NewApplicationService(db, companyClient))

	// Запуск сервиса
	log.Info().Int("port", cfg.ApplicationService.Port).Msg("Application service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("Failed to start grpc server")
	}
}
