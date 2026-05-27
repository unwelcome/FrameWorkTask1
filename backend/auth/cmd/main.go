package main

import (
	"fmt"
	"net"

	"github.com/rs/zerolog/log"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/config"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/services"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/interceptors"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/logger"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.NewConfig()

	loggerConf, httpLogger := logger.Setup(cfg.Log.Path, cfg.Log.ConsoleOut)
	log.Logger = *loggerConf

	// Загружаем приватный ключ из PEM-файла
	privateKey, err := utils.LoadPrivateKey(cfg.JWT.PrivateKeyPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", cfg.JWT.PrivateKeyPath).Msg("failed to load JWT private key")
	}

	db := postgresDB.NewDatabaseInstance(cfg.Postgres.ConnectionString())
	cache := redisDB.NewCacheInstance(cfg.Redis.Options(), cfg.JWT.RefreshTokenLifetime, cfg.Redis.Prefix)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start tcp server")
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptors.NewLoggingInterceptor(*httpLogger)),
	)
	auth_proto.RegisterAuthServiceServer(grpcServer, services.NewAuthService(
		db, cache,
		privateKey,
		cfg.JWT.AccessTokenLifetime,
		cfg.JWT.RefreshTokenLifetime,
	))

	log.Info().Int("port", cfg.Port).Msg("auth service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
