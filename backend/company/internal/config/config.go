package config

import (
	sharedConfig "github.com/unwelcome/FrameWorkTask1/backend/shared/config"
)

type Config struct {
	Port        int
	MetricsPort int
	Log         LogConfig
	Postgres    sharedConfig.PostgresConfig
	Redis       sharedConfig.RedisConfig
}

type LogConfig struct {
	Path       string
	ConsoleOut bool
}

func NewConfig() *Config {
	return &Config{
		Port:        sharedConfig.MustParseInt("COMPANY_SERVICE_PORT"),
		MetricsPort: sharedConfig.ParseIntOrDefault("METRICS_PORT", 2112),
		Log: LogConfig{
			Path:       sharedConfig.MustGetEnv("LOG_PATH"),
			ConsoleOut: sharedConfig.MustParseBool("LOG_CONSOLE_OUT"),
		},
		Postgres: sharedConfig.NewPostgresConfig(),
		Redis:    sharedConfig.NewRedisConfig(),
	}
}
