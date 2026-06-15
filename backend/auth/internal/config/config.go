package config

import (
	"time"

	sharedConfig "github.com/unwelcome/FrameWorkTask1/backend/shared/config"
)

type Config struct {
	Port     int
	AppEnv   string
	Log      LogConfig
	Postgres sharedConfig.PostgresConfig
	Redis    sharedConfig.RedisConfig
	RabbitMQ sharedConfig.RabbitMQConfig
	JWT      JWTConfig
	Password PasswordConfig
}

// PasswordConfig ограничивает одновременные вычисления Argon2 (защита от resource-exhaustion DoS).
// Потолок памяти на хеширование: MaxConcurrentHashes × 64 MiB.
type PasswordConfig struct {
	MaxConcurrentHashes int
	AcquireTimeout      time.Duration
}

type LogConfig struct {
	Path       string
	ConsoleOut bool
}

type JWTConfig struct {
	PrivateKeyPath       string
	AccessTokenLifetime  time.Duration
	RefreshTokenLifetime time.Duration
}

func NewConfig() *Config {
	return &Config{
		Port:   sharedConfig.MustParseInt("AUTH_SERVICE_PORT"),
		AppEnv: sharedConfig.GetEnvOrDefault("APP_ENV", "production"),
		Log: LogConfig{
			Path:       sharedConfig.MustGetEnv("LOG_PATH"),
			ConsoleOut: sharedConfig.MustParseBool("LOG_CONSOLE_OUT"),
		},
		Postgres: sharedConfig.NewPostgresConfig(),
		Redis:    sharedConfig.NewRedisConfig(),
		RabbitMQ: sharedConfig.NewRabbitMQConfig(),
		JWT: JWTConfig{
			PrivateKeyPath:       sharedConfig.MustGetEnv("JWT_PRIVATE_KEY_PATH"),
			AccessTokenLifetime:  sharedConfig.MustParseDuration("ACCESS_TOKEN_LIFETIME"),
			RefreshTokenLifetime: sharedConfig.MustParseDuration("REFRESH_TOKEN_LIFETIME"),
		},
		Password: PasswordConfig{
			MaxConcurrentHashes: sharedConfig.ParseIntOrDefault("MAX_CONCURRENT_HASHES", 8),
			AcquireTimeout:      sharedConfig.ParseDurationOrDefault("HASH_ACQUIRE_TIMEOUT", 3*time.Second),
		},
	}
}
