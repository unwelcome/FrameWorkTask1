package config

import (
	"time"

	sharedConfig "github.com/unwelcome/FrameWorkTask1/backend/shared/config"
)

type Config struct {
	Port     int
	Log      LogConfig
	Postgres sharedConfig.PostgresConfig
	Redis    sharedConfig.RedisConfig
	JWT      JWTConfig
}

type LogConfig struct {
	Path       string
	ConsoleOut bool
}

type JWTConfig struct {
	Secret               string
	AccessTokenLifetime  time.Duration
	RefreshTokenLifetime time.Duration
}

func NewConfig() *Config {
	return &Config{
		Port: sharedConfig.MustParseInt("AUTH_SERVICE_PORT"),
		Log: LogConfig{
			Path:       sharedConfig.MustGetEnv("LOG_PATH"),
			ConsoleOut: sharedConfig.MustParseBool("LOG_CONSOLE_OUT"),
		},
		Postgres: sharedConfig.NewPostgresConfig(),
		Redis:    sharedConfig.NewRedisConfig(),
		JWT: JWTConfig{
			Secret:               sharedConfig.MustGetEnv("JWT_SECRET"),
			AccessTokenLifetime:  sharedConfig.MustParseDuration("ACCESS_TOKEN_LIFETIME"),
			RefreshTokenLifetime: sharedConfig.MustParseDuration("REFRESH_TOKEN_LIFETIME"),
		},
	}
}
