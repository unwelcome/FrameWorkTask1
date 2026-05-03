package redisDB

import (
	"github.com/redis/go-redis/v9"
	sharedRedis "github.com/unwelcome/FrameWorkTask1/backend/shared/redis"
)

type CacheRepository struct {
	Company CompanyRepository
}

func NewCacheInstance(connectOptions *redis.Options) *CacheRepository {
	rdb := sharedRedis.Connect(connectOptions)

	return &CacheRepository{
		Company: NewCompanyRepository(rdb),
	}
}
