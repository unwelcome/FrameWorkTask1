package main

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/auth/api/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/config"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/services"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/logger"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.NewConfig()

	loggerConf := logger.Setup(cfg.Log.Path, cfg.Log.ConsoleOut)
	log.Logger = *loggerConf

	db := postgresDB.NewDatabaseInstance(cfg.Postgres.ConnectionString())
	cache := redisDB.NewCacheInstance(cfg.Redis.Options(), cfg.JWT.RefreshTokenLifetime)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start tcp server")
	}

	grpcServer := grpc.NewServer()
	auth_proto.RegisterAuthServiceServer(grpcServer, services.NewAuthService(
		db, cache,
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenLifetime,
		cfg.JWT.RefreshTokenLifetime,
	))

	log.Info().Int("port", cfg.Port).Msg("auth service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
