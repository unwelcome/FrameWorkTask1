package main

import (
	"fmt"
	"net"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog/log"
	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/config"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/company/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/company/internal/services"
	company_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/company/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/interceptors"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/logger"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/metrics"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.NewConfig()

	loggerConf, httpLogger := logger.Setup(cfg.Log.Path, cfg.Log.ConsoleOut)
	log.Logger = *loggerConf

	db := postgresDB.NewDatabaseInstance(cfg.Postgres.ConnectionString())
	cache := redisDB.NewCacheInstance(cfg.Redis.Options(), cfg.Redis.Prefix)

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
	company_proto.RegisterCompanyServiceServer(grpcServer, services.NewCompanyService(db, cache))

	grpcprom.Register(grpcServer)

	metrics.StartServer(cfg.MetricsPort)

	log.Info().Int("port", cfg.Port).Msg("company service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
