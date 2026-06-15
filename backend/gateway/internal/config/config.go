package config

import (
	"fmt"

	sharedConfig "github.com/unwelcome/FrameWorkTask1/backend/shared/config"
)

type Config struct {
	Port      int
	AppEnv    string
	Log       LogConfig
	JWT       JWTConfig
	GeoIP     GeoIPConfig
	Redis     sharedConfig.RedisConfig
	RateLimit RateLimitConfig
	Auth      ServiceAddress
	Company   ServiceAddress
	App       ServiceAddress
}

type RateLimitConfig struct {
	Password RateLimitTierConfig
	Code     RateLimitTierConfig
	User     RateLimitTierConfig
}

// RateLimitTierConfig описывает token bucket одного тира.
// Capacity — размер корзины (максимальный burst).
// RefillPerSec — скорость пополнения, токенов в секунду (устойчивый rate).
type RateLimitTierConfig struct {
	Capacity     int
	RefillPerSec float64
}

// GeoIPConfig содержит пути к базам данных MaxMind GeoLite2.
// Оба поля опциональны: если файл не указан или не найден,
// соответствующие поля сессии останутся пустыми.
type GeoIPConfig struct {
	CityDBPath string // Путь к GeoLite2-City.mmdb
	ASNDBPath  string // Путь к GeoLite2-ASN.mmdb
}

type LogConfig struct {
	Path       string
	ConsoleOut bool
}

type JWTConfig struct {
	PublicKeyPath string
}

type ServiceAddress struct {
	Host string
	Port int
}

func (s ServiceAddress) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func NewConfig() *Config {
	return &Config{
		Port:   sharedConfig.MustParseInt("GATEWAY_PORT"),
		AppEnv: sharedConfig.GetEnvOrDefault("APP_ENV", "production"),
		Log: LogConfig{
			Path:       sharedConfig.MustGetEnv("LOG_PATH"),
			ConsoleOut: sharedConfig.MustParseBool("LOG_CONSOLE_OUT"),
		},
		JWT: JWTConfig{
			PublicKeyPath: sharedConfig.MustGetEnv("JWT_PUBLIC_KEY_PATH"),
		},
		GeoIP: GeoIPConfig{
			CityDBPath: sharedConfig.GetEnvOrDefault("GEOIP_CITY_DB_PATH", ""),
			ASNDBPath:  sharedConfig.GetEnvOrDefault("GEOIP_ASN_DB_PATH", ""),
		},
		Redis: sharedConfig.NewRedisConfig(),
		RateLimit: RateLimitConfig{
			Password: RateLimitTierConfig{
				Capacity:     sharedConfig.ParseIntOrDefault("RATE_LIMIT_PASSWORD_CAPACITY", 50),
				RefillPerSec: sharedConfig.ParseFloatOrDefault("RATE_LIMIT_PASSWORD_REFILL", 1),
			},
			Code: RateLimitTierConfig{
				Capacity:     sharedConfig.ParseIntOrDefault("RATE_LIMIT_CODE_CAPACITY", 100),
				RefillPerSec: sharedConfig.ParseFloatOrDefault("RATE_LIMIT_CODE_REFILL", 2),
			},
			User: RateLimitTierConfig{
				Capacity:     sharedConfig.ParseIntOrDefault("RATE_LIMIT_USER_CAPACITY", 100),
				RefillPerSec: sharedConfig.ParseFloatOrDefault("RATE_LIMIT_USER_REFILL", 1),
			},
		},
		Auth: ServiceAddress{
			Host: sharedConfig.MustGetEnv("AUTH_SERVICE_HOST"),
			Port: sharedConfig.MustParseInt("AUTH_SERVICE_PORT"),
		},
		Company: ServiceAddress{
			Host: sharedConfig.MustGetEnv("COMPANY_SERVICE_HOST"),
			Port: sharedConfig.MustParseInt("COMPANY_SERVICE_PORT"),
		},
		App: ServiceAddress{
			Host: sharedConfig.MustGetEnv("APPLICATION_SERVICE_HOST"),
			Port: sharedConfig.MustParseInt("APPLICATION_SERVICE_PORT"),
		},
	}
}
