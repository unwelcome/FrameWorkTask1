package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfig(l *slog.Logger) *Config {
	// Загружаем .env файл (игнорируем ошибку если файла нет)
	_ = godotenv.Load("../.env")

	//Для локального запуска (IDE перезаписывает IS_DOCKER=false)
	dbHost := "localhost"
	//Для запуска через Docker (IS_DOCKER=true из .env файла)
	if getEnv("IS_DOCKER", "") == "true" {
		dbHost = getEnv("POSTGRES_HOST", "postgres")
	}

	dbPort := getEnv("POSTGRES_PORT", "5432")
	dbUser := getEnv("POSTGRES_USER", "postgres")
	dbPassword := getEnv("POSTGRES_PASSWORD", "postgres")
	dbName := getEnv("POSTGRES_DB", "app")

	l.Info(
		"ENV CONFIGURATION",
		"DBHOST", dbHost,
		"DBPORT", dbPort,
		"DBUSER", dbUser,
		"DBPASSWORD", dbPassword,
		"DBNAME", dbName)

	return &Config{
		DBHost:     dbHost,
		DBPort:     dbPort,
		DBUser:     dbUser,
		DBPassword: dbPassword,
		DBName:     dbName,
	}
}

func (c *Config) DBConnString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
