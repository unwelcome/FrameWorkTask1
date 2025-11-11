package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App                AppConfig                `yaml:"app"`
	Gateway            GatewayConfig            `yaml:"gateway"`
	AuthService        AuthServiceConfig        `yaml:"auth_service"`
	ApplicationService ApplicationServiceConfig `yaml:"application_service"`
	Db                 Database                 `yaml:"database"`
	Cache              Cache                    `yaml:"cache"`
	S3                 S3                       `yaml:"s3"`
}

type AppConfig struct {
	ProductionType          string `env:"PRODUCTION_TYPE"`
	JWTSecret               string `yaml:"jwt_secret"`
	LogPath                 string `yaml:"log_path"`
	LogConsoleOut           bool   `yaml:"log_console_out"`
	AccessTokenLifetimeStr  string `yaml:"access_token_lifetime"`
	AccessTokenLifetime     time.Duration
	RefreshTokenLifetimeStr string `yaml:"refresh_token_lifetime"`
	RefreshTokenLifetime    time.Duration
}

type GatewayConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type AuthServiceConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	DBUser     string `yaml:"db_user"`
	DBPassword string `yaml:"db_password"`
	DBName     string `yaml:"db_name"`
	CacheDB    int    `yaml:"cache_db"`
	S3Bucket   string `yaml:"s3_bucket"`
}

type ApplicationServiceConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	DBUser     string `yaml:"db_user"`
	DBPassword string `yaml:"db_password"`
	DBName     string `yaml:"db_name"`
	CacheDB    int    `yaml:"cache_db"`
	S3Bucket   string `yaml:"s3_bucket"`
}

type Database struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type Cache struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type S3 struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func NewConfig() *Config {
	config := &Config{}

	// Определяем среду запуска
	prodType := os.Getenv("PRODUCTION_TYPE")

	// Если пусто - загружаем .env файл
	if prodType == "" {
		if err := godotenv.Load("../../.env"); err != nil {
			panic(fmt.Errorf(".env file not found"))
		}
		prodType = os.Getenv("PRODUCTION_TYPE")
	}

	// Задаем путь к файлу конфигурации
	configPath := "config/config.yaml" // PRODUCTION_TYPE=prod
	if prodType == "dev" {
		configPath = "../config.yaml" // PRODUCTION_TYPE=dev
	}

	fmt.Printf("ProductionType: %s\n", prodType)
	fmt.Printf("ConfigPath: %s\n", configPath)

	// Загружаем конфиг
	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Errorf("failed read config file: %w", err))
	}

	// Инициализируем конфиг
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(fmt.Errorf("failed parse config file: %w", err))
	}

	// Устанавливаем ProductionType
	config.App.ProductionType = prodType

	// Переводим время из string в time.Duration
	err = config.parseDurations()
	if err != nil {
		panic(fmt.Errorf("failed parse duration: %w", err))
	}

	return config
}

func (config *Config) parseDurations() error {
	var err error

	// Парсим время жизни access токена
	config.App.AccessTokenLifetime, err = time.ParseDuration(config.App.AccessTokenLifetimeStr)
	if err != nil {
		return fmt.Errorf("invalid access token lifetime: %w", err)
	}

	// Парсим время жизни refresh токена
	config.App.RefreshTokenLifetime, err = time.ParseDuration(config.App.RefreshTokenLifetimeStr)
	if err != nil {
		return fmt.Errorf("invalid refresh token lifetime: %w", err)
	}

	return nil
}

func (config *Config) Print() {
	hideCredentials := func(credential string) string {
		return strings.Repeat("*", len(credential))
	}

	fmt.Printf("=== CONFIG ===\n")

	fmt.Printf("=== App ===\n")
	fmt.Printf("ProductionType: %s\n", config.App.ProductionType)
	fmt.Printf("JWTSecret: %s\n", hideCredentials(config.App.JWTSecret))
	fmt.Printf("LogPath: %s\n", config.App.LogPath)
	fmt.Printf("LogConsoleOut: %v\n", config.App.LogConsoleOut)
	fmt.Printf("AccessTokenLifetime: %v\n", config.App.AccessTokenLifetime)
	fmt.Printf("RefreshTokenLifetime: %v\n", config.App.RefreshTokenLifetime)

	fmt.Printf("=== Gateway ===\n")
	fmt.Printf("Host: %s\n", config.Gateway.Host)
	fmt.Printf("Port: %d\n", config.Gateway.Port)

	fmt.Printf("=== Auth service ===\n")
	fmt.Printf("Port: %d\n", config.AuthService.Port)
	fmt.Printf("DBUser: %s\n", config.AuthService.DBUser)
	fmt.Printf("DBPassword: %s\n", hideCredentials(config.AuthService.DBPassword))
	fmt.Printf("DBName: %s\n", config.AuthService.DBName)
	fmt.Printf("CacheDB: %d\n", config.AuthService.CacheDB)
	fmt.Printf("S3Bucket: %s\n", config.AuthService.S3Bucket)

	fmt.Printf("=== Application service ===\n")
	fmt.Printf("Port: %d\n", config.ApplicationService.Port)
	fmt.Printf("DBUser: %s\n", config.ApplicationService.DBUser)
	fmt.Printf("DBPassword: %s\n", hideCredentials(config.ApplicationService.DBPassword))
	fmt.Printf("DBName: %s\n", config.ApplicationService.DBName)
	fmt.Printf("CacheDB: %d\n", config.ApplicationService.CacheDB)
	fmt.Printf("S3Bucket: %s\n", config.ApplicationService.S3Bucket)

	fmt.Printf("=== Database ===\n")
	fmt.Printf("Host: %s\n", config.Db.Host)
	fmt.Printf("Port: %d\n", config.Db.Port)

	fmt.Printf("=== Cache ===\n")
	fmt.Printf("Host: %s\n", config.Cache.Host)
	fmt.Printf("Port: %d\n", config.Cache.Port)
	fmt.Printf("Password: %s\n", hideCredentials(config.Cache.Password))

	fmt.Printf("=== S3 ===\n")
	fmt.Printf("Host: %s\n", config.S3.Host)
	fmt.Printf("Port: %d\n", config.S3.Port)
}
