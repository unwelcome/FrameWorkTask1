package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/unwelcome/iqjtest/database/minio"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppInternalHost string
	AppInternalPort string
	AppPublicHost   string
	AppPublicPort   string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	CacheHost     string
	CachePort     string
	CacheUser     string
	CachePassword string
	CacheDBName   string

	S3Host     string
	S3Port     string
	S3User     string
	S3Password string
	S3UseSSL   bool
	S3Buckets  map[string]*miniodb.Bucket

	BCryptCost           int
	JWTSecret            string
	AccessTokenLifetime  time.Duration
	RefreshTokenLifetime time.Duration

	Timeouts struct {
		Middleware  time.Duration
		Request     time.Duration
		FileRequest time.Duration
	}
}

func LoadConfig(l zerolog.Logger) *Config {
	// Загружаем .env файл (игнорируем ошибку если файла нет)
	_ = godotenv.Load("../.env")

	// Создаем экземпляр конфига
	cfg := &Config{}

	// Инициализируем данные из env файла для App
	cfg.AppInternalHost = "localhost"
	cfg.AppInternalPort = getEnv("BACKEND_INTERNAL_PORT", "8080")
	cfg.AppPublicHost = "localhost"
	cfg.AppPublicPort = getEnv("BACKEND_PUBLIC_PORT", "8080")

	// Инициализируем данные из env файла для основной DB
	cfg.DBHost = "localhost"
	cfg.DBPort = getEnv("POSTGRES_PORT", "5432")
	cfg.DBUser = getEnv("POSTGRES_USER", "postgres")
	cfg.DBPassword = getEnv("POSTGRES_PASSWORD", "postgres")
	cfg.DBName = getEnv("POSTGRES_DB", "app")

	// Инициализируем данные из env файла для кеширования
	cfg.CacheHost = "localhost"
	cfg.CachePort = getEnv("REDIS_PORT", "6379")
	cfg.CacheUser = getEnv("REDIS_USER", "redis")
	cfg.CachePassword = getEnv("REDIS_PASSWORD", "redis")
	cfg.CacheDBName = getEnv("REDIS_DB", "0")

	// Инициализируем данные из env файла для s3 хранилища
	cfg.S3Host = "localhost"
	cfg.S3Port = getEnv("MINIO_PORT", "9000")
	cfg.S3User = getEnv("MINIO_USER", "minio")
	cfg.S3Password = getEnv("MINIO_PASSWORD", "minio")
	cfg.S3UseSSL = getEnvBool("MINIO_SSL", false)

	// Для запуска через Docker
	if getEnv("IS_DOCKER", "") == "true" {
		cfg.AppInternalHost = getEnv("BACKEND_INTERNAL_HOST", "0.0.0.0")
		cfg.AppPublicHost = getEnv("BACKEND_PUBLIC_HOST", "localhost")
		cfg.DBHost = getEnv("POSTGRES_HOST", "postgres")
		cfg.CacheHost = getEnv("REDIS_HOST", "redis")
		cfg.S3Host = getEnv("MINIO_HOST", "minio")
	}

	// Устанавливаем стойкость шифрования пароля
	cfg.BCryptCost = 10

	// Инициализируем jwt секрет
	cfg.JWTSecret = getEnv("JWT_SECRET", "ultra-secret-key")
	cfg.AccessTokenLifetime = 5 * time.Minute
	cfg.RefreshTokenLifetime = 30 * 24 * time.Hour

	// Декларируем S3 бакеты
	cfg.S3Buckets = map[string]*miniodb.Bucket{
		"catPhotoBucket": &miniodb.Bucket{Name: "cat-photo-bucket", IsOpen: true},
	}

	// Устанавливаем время выполнения запросов
	cfg.Timeouts.Middleware = time.Second * 5
	cfg.Timeouts.Request = time.Second * 5
	cfg.Timeouts.FileRequest = time.Second * 30

	l.Trace().Str("AppInternalHost", cfg.AppInternalHost).Str("AppInternalPort", cfg.AppInternalPort).Msg("App internal config")
	l.Trace().Str("AppPublicHost", cfg.AppPublicHost).Str("AppPublicPort", cfg.AppPublicPort).Msg("App public config")
	l.Trace().Str("DBHost", cfg.DBHost).Str("DBPort", cfg.DBPort).Msg("Postgres config")
	l.Trace().Str("CacheHost", cfg.CacheHost).Str("CachePort", cfg.CachePort).Msg("Redis config")
	l.Trace().Str("S3Host", cfg.S3Host).Str("S3Port", cfg.S3Port).Msg("Minio config")

	// Возвращаем экземпляр конфига
	return cfg
}

func (c *Config) GetAppInternalAddress() string {
	return fmt.Sprintf("%s:%s", c.AppInternalHost, c.AppInternalPort)
}

func (c *Config) GetDBConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func (c *Config) GetCacheConnString() string {
	//"redis://<user>:<pass>@localhost:6379/<db>"
	return fmt.Sprintf("redis://%s:%s@%s:%s/%s", c.CacheUser, c.CachePassword, c.CacheHost, c.CachePort, c.CacheDBName)
}

func (c *Config) S3ConnConfig() *miniodb.ConnectConfig {
	return &miniodb.ConnectConfig{
		BackendEndpoint: fmt.Sprintf("%s:%s", c.S3Host, c.S3Port),
		PublicEndpoint:  fmt.Sprintf("%s:%s", c.AppPublicHost, c.S3Port),
		Username:        c.S3User,
		Password:        c.S3Password,
		UseSSL:          c.S3UseSSL,
	}
}

func (c *Config) GetS3Buckets() []*miniodb.Bucket {
	var buckets []*miniodb.Bucket
	for _, bucket := range c.S3Buckets {
		buckets = append(buckets, bucket)
	}
	return buckets
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		result, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return result
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		result, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return result
	}
	return defaultValue
}
