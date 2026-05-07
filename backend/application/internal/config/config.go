package config

import (
	"fmt"

	sharedConfig "github.com/unwelcome/FrameWorkTask1/backend/shared/config"
)

type Config struct {
	Port           int
	Log            LogConfig
	Postgres       sharedConfig.PostgresConfig
	Mongo          sharedConfig.MongoDBConfig
	CompanyService ServiceAddress
}

type LogConfig struct {
	Path       string
	ConsoleOut bool
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
		Port: sharedConfig.MustParseInt("APPLICATION_SERVICE_PORT"),
		Log: LogConfig{
			Path:       sharedConfig.MustGetEnv("LOG_PATH"),
			ConsoleOut: sharedConfig.MustParseBool("LOG_CONSOLE_OUT"),
		},
		Postgres: sharedConfig.NewPostgresConfig(),
		Mongo:    sharedConfig.NewMongoDBConfig(),
		CompanyService: ServiceAddress{
			Host: sharedConfig.MustGetEnv("COMPANY_SERVICE_HOST"),
			Port: sharedConfig.MustParseInt("COMPANY_SERVICE_PORT"),
		},
	}
}
