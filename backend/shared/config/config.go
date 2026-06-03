package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// ─── Postgres ─────────────────────────────────────────────────────────────────

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
}

func NewPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     MustGetEnv("POSTGRES_HOST"),
		Port:     MustParseInt("POSTGRES_PORT"),
		User:     MustGetEnv("POSTGRES_USER"),
		Password: MustGetEnv("POSTGRES_PASSWORD"),
		DB:       MustGetEnv("POSTGRES_DB"),
	}
}

func (c PostgresConfig) ConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.DB,
	)
}

// ─── Redis ────────────────────────────────────────────────────────────────────

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	Prefix   string
}

func NewRedisConfig() RedisConfig {
	return RedisConfig{
		Host:     MustGetEnv("REDIS_HOST"),
		Port:     MustParseInt("REDIS_PORT"),
		Password: MustGetEnv("REDIS_PASSWORD"),
		Prefix:   GetEnvOrDefault("REDIS_PREFIX", ""),
	}
}

func (c RedisConfig) Options() *redis.Options {
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password: c.Password,
	}
}

// ─── Minio ────────────────────────────────────────────────────────────────────

type MinioConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Bucket   string
	SSL      bool
}

func NewMinioConfig() MinioConfig {
	return MinioConfig{
		Host:     MustGetEnv("MINIO_HOST"),
		Port:     MustParseInt("MINIO_PORT"),
		User:     MustGetEnv("MINIO_ROOT_USER"),
		Password: MustGetEnv("MINIO_ROOT_PASSWORD"),
		Bucket:   MustGetEnv("MINIO_BUCKET"),
		SSL:      MustParseBool("MINIO_SSL"),
	}
}

// ─── MongoDB ────────────────────────────────────────────────────────────────────

type MongoDBConfig struct {
	Host         string
	Port         int
	User         string
	RootUser     string
	Password     string
	RootPassword string
	DbName       string
}

func NewMongoDBConfig() MongoDBConfig {
	return MongoDBConfig{
		Host:         MustGetEnv("MONGO_HOST"),
		Port:         MustParseInt("MONGO_PORT"),
		User:         MustGetEnv("MONGO_USER"),
		RootUser:     MustGetEnv("MONGO_INITDB_ROOT_USERNAME"),
		Password:     MustGetEnv("MONGO_PASSWORD"),
		RootPassword: MustGetEnv("MONGO_INITDB_ROOT_PASSWORD"),
		DbName:       MustGetEnv("MONGO_DB"),
	}
}

func (c MongoDBConfig) RootConnectionString() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/admin", c.RootUser, c.RootPassword, c.Host, c.Port)
}

func (c MongoDBConfig) ConnectionString() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.DbName)
}

// ─── RabbitMQ ───────────────────────────────────────────────────────────────────

type RabbitMQConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

func NewRabbitMQConfig() RabbitMQConfig {
	return RabbitMQConfig{
		Host:     MustGetEnv("RABBITMQ_HOST"),
		Port:     MustParseInt("RABBITMQ_PORT"),
		User:     MustGetEnv("RABBITMQ_USER"),
		Password: MustGetEnv("RABBITMQ_PASSWORD"),
	}
}

func (c RabbitMQConfig) ConnectionString() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/", c.User, c.Password, c.Host, c.Port)
}

// ─── Утилиты ──────────────────────────────────────────────────────────────────

func MustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func GetEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func MustParseInt(key string) int {
	v := MustGetEnv(key)
	n, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("environment variable %q must be an integer, got %q", key, v))
	}
	return n
}

func MustParseBool(key string) bool {
	v := MustGetEnv(key)
	b, err := strconv.ParseBool(v)
	if err != nil {
		panic(fmt.Sprintf("environment variable %q must be a boolean (true/false), got %q", key, v))
	}
	return b
}

func MustParseDuration(key string) time.Duration {
	v := MustGetEnv(key)
	d, err := time.ParseDuration(v)
	if err != nil {
		panic(fmt.Sprintf("environment variable %q must be a duration (e.g. 5m, 720h), got %q", key, v))
	}
	return d
}
