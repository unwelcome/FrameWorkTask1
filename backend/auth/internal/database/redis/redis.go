package redisDB

import (
	"time"

	"github.com/redis/go-redis/v9"
	sharedRedis "github.com/unwelcome/FrameWorkTask1/backend/shared/redis"
)

type CacheRepository struct {
	Auth AuthRepository
}

func NewCacheInstance(connectOptions *redis.Options, refreshTokenTTL time.Duration, prefix string) *CacheRepository {
	rdb := sharedRedis.Connect(connectOptions)

	return &CacheRepository{
		Auth: NewAuthRepository(rdb, refreshTokenTTL, prefix),
	}
}
