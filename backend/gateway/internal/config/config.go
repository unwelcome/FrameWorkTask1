package config

import (
	"fmt"

	sharedConfig "github.com/unwelcome/FrameWorkTask1/backend/shared/config"
)

type Config struct {
	Port    int
	AppEnv  string
	Log     LogConfig
	JWT     JWTConfig
	Auth    ServiceAddress
	Company ServiceAddress
	App     ServiceAddress
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
