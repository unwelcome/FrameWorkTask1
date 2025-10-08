package redisdb

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func Connect(ctx context.Context, connectString string, l zerolog.Logger) *redis.Client {
	opt, err := redis.ParseURL(connectString)
	if err != nil {
		l.Fatal().Err(err).Msg("failed to parse redis connect string")
	}

	rdb := redis.NewClient(opt)

	if err = rdb.Ping(ctx).Err(); err != nil {
		l.Fatal().Err(err).Msg("failed to connect to redis server")
		return nil
	}

	l.Trace().Msg("Successfully connected to Redis!")
	return rdb
}
