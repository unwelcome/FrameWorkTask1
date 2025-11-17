package redisDB

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type CacheRepository struct {
	Company CompanyRepository
}

func NewCacheInstance(connectOptions *redis.Options) *CacheRepository {
	// Подключаемся к Redis
	rdb := redis.NewClient(connectOptions)

	// Проверка подключения
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect to redis")
		return nil
	}

	// Создаем структуру redis репозиториев
	cacheRepository := &CacheRepository{}

	// Создаем репозитории
	cacheRepository.Company = NewCompanyRepository(rdb)

	return cacheRepository
}
