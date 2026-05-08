package main

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/application/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/config"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/services"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/company/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.NewConfig()

	loggerConf := logger.Setup(cfg.Log.Path, cfg.Log.ConsoleOut)
	log.Logger = *loggerConf

	db := postgresDB.NewDatabaseInstance(cfg.Postgres.ConnectionString())

	companyConn, err := grpc.NewClient(cfg.CompanyService.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal().Err(err).Str("addr", cfg.CompanyService.Addr()).Msg("failed to connect to company service")
	}
	defer companyConn.Close()

	companyClient := company_proto.NewCompanyServiceClient(companyConn)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start tcp server")
	}

	grpcServer := grpc.NewServer()
	application_proto.RegisterApplicationServiceServer(grpcServer, services.NewApplicationService(db, companyClient))

	log.Info().Int("port", cfg.Port).Msg("application service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
