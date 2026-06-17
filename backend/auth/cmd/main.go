package main

import (
	"context"
	"fmt"
	"net"
	"os/signal"
	"syscall"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/rs/zerolog/log"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/config"
	postgresDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/postgres"
	redisDB "github.com/unwelcome/FrameWorkTask1/backend/auth/internal/database/redis"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/messaging"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/internal/services"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/password"
	"github.com/unwelcome/FrameWorkTask1/backend/auth/pkg/utils"
	auth_proto "github.com/unwelcome/FrameWorkTask1/backend/contracts/auth/generated"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/interceptors"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/logger"
	"github.com/unwelcome/FrameWorkTask1/backend/shared/metrics"
	"google.golang.org/grpc"
)

func main() {
	cfg := config.NewConfig()

	loggerConf, httpLogger := logger.Setup(cfg.Log.Path, cfg.Log.ConsoleOut)
	log.Logger = *loggerConf

	// Ограничиваем одновременные вычисления Argon2 (защита от resource-exhaustion DoS)
	password.Setup(cfg.Password.MaxConcurrentHashes, cfg.Password.AcquireTimeout)

	// Загружаем приватный ключ из PEM-файла
	privateKey, err := utils.LoadPrivateKey(cfg.JWT.PrivateKeyPath)
	if err != nil {
		log.Fatal().Err(err).Str("path", cfg.JWT.PrivateKeyPath).Msg("failed to load JWT private key")
	}

	db := postgresDB.NewDatabaseInstance(cfg.Postgres.ConnectionString())
	cache := redisDB.NewCacheInstance(cfg.Redis.Options(), cfg.JWT.RefreshTokenLifetime, cfg.Redis.Prefix)
	rabbitMQ := messaging.NewPublisher(cfg.RabbitMQ.ConnectionString())

	// Контекст для graceful shutdown, отменяется по SIGINT / SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Фоновая горутина анонимизации удалённых аккаунтов
	go services.StartCleanupWorker(ctx, db)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start tcp server")
	}

	// Включаем гистограмму latency (выключена по умолчанию из-за кардинальности)
	grpcprom.EnableHandlingTimeHistogram()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcprom.UnaryServerInterceptor,
			interceptors.NewLoggingInterceptor(*httpLogger),
		),
		grpc.StreamInterceptor(grpcprom.StreamServerInterceptor),
	)
	auth_proto.RegisterAuthServiceServer(grpcServer, services.NewAuthService(
		db, cache, rabbitMQ,
		privateKey,
		cfg.JWT.AccessTokenLifetime,
		cfg.JWT.RefreshTokenLifetime,
		cfg.AppEnv,
	))

	// Инициализируем метрики для всех зарегистрированных методов
	grpcprom.Register(grpcServer)

	metrics.StartServer(cfg.MetricsPort)

	log.Info().Int("port", cfg.Port).Msg("auth service started")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal().Err(err).Msg("failed to serve grpc")
	}
}
