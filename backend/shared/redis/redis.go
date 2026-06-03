package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	maxRetries = 10
	retryDelay = 3 * time.Second
)

func Connect(options *redis.Options) *redis.Client {
	rdb := redis.NewClient(options)

	var err error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = rdb.Ping(context.Background()).Err()
		if err == nil {
			break
		}
		log.Warn().
			Err(err).
			Int("attempt", attempt).
			Int("max", maxRetries).
			Msgf("failed to ping redis, retrying in %s...", retryDelay)
		time.Sleep(retryDelay)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("failed to ping redis after all retries")
	}

	return rdb
}
