package main

import (
	"fmt"
	"net"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog/log"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/config"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/application/internal/database/postgres"
	"github.com/unwelcome/FrameWorkTask1/backend/application/internal/services"
	application_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/application/generated"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/interceptors"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/logger"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	cfg := config.NewConfig()

	loggerConf, httpLogger := logger.Setup(cfg.Log.Path, cfg.Log.ConsoleOut)
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

	grpcprom.EnableHandlingTimeHistogram()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcprom.UnaryServerInterceptor,
			interceptors.NewLoggingInterceptor(*httpLogger),
		),
		grpc.StreamInterceptor(grpcprom.StreamServerInterceptor),
	)
	application_proto.RegisterApplicationServiceServer(grpcServer, services.NewApplicationService(db, companyClient))

	grpcprom.Register(grpcServer)

	metrics.StartServer(cfg.MetricsPort)

	log.Info().Int("port", cfg.Port).Msg("application service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
