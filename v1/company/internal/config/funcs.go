package config

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

// GetDBConnectionString Строка подключения к Postgres для company сервиса
func (config *Config) GetDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Db.Host, config.Db.Port, config.CompanyService.DBUser, config.CompanyService.DBPassword, config.CompanyService.DBName)
}

// GetCacheConnectionOptions Конфиг для подключения к Redis
func (config *Config) GetCacheConnectionOptions() *redis.Options {
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Cache.Host, config.Cache.Port),
		Password: config.Cache.Password,
		DB:       config.CompanyService.CacheDB,
	}
}
