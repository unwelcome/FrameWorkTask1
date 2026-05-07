package redisDB

import (
	"context"

	"github.com/redis/go-redis/v9"
	sharedRedis "github.com/unwelcome/FrameWorkTask1/backend/shared/redis"
)

type CacheRepository struct {
	Company CompanyRepository
	rdb     *redis.Client
}

func (r *CacheRepository) Ping(ctx context.Context) error {
	return r.rdb.Ping(ctx).Err()
}

func NewCacheInstance(connectOptions *redis.Options, prefix string) *CacheRepository {
	rdb := sharedRedis.Connect(connectOptions)

	return &CacheRepository{
		Company: NewCompanyRepository(rdb, prefix),
		rdb:     rdb,
	}
}
