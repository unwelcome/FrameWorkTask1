package redisDB

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	sharedRedis "github.com/unwelcome/FrameWorkTask1/backend/shared/redis"
)

type CacheRepository struct {
	Auth         AuthRepository
	Verification VerificationRepository
	rdb          *redis.Client
}

func (r *CacheRepository) Ping(ctx context.Context) error {
	return r.rdb.Ping(ctx).Err()
}

func NewCacheInstance(connectOptions *redis.Options, refreshTokenTTL time.Duration, prefix string) *CacheRepository {
	rdb := sharedRedis.Connect(connectOptions)

	return &CacheRepository{
		Auth:         NewAuthRepository(rdb, refreshTokenTTL, prefix),
		Verification: NewVerificationRepository(rdb, prefix),
		rdb:          rdb,
	}
}
