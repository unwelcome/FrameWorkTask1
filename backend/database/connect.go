package database

import (
	"context"
	"database/sql"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/unwelcome/iqjtest/internal/config"

	"github.com/unwelcome/iqjtest/database/minio"
	"github.com/unwelcome/iqjtest/database/postgresql"
	"github.com/unwelcome/iqjtest/database/redis"
)

func ConnectToDatabases(cfg *config.Config, logger zerolog.Logger) (*sql.DB, *redis.Client, *minio.Client) {

	// Подключение к postgresql
	Postgres := postgresdb.Connect(cfg.GetDBConnString(), logger)

	// Подключение к redis
	Redis := redisdb.Connect(context.Background(), cfg.GetCacheConnString(), logger)

	// Подключение к MinIO
	Minio := miniodb.InitMinio(cfg.S3ConnConfig(), logger, cfg.GetS3Buckets())

	return Postgres, Redis, Minio
}
