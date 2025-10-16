package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/rs/zerolog"
	"time"
)

type Config struct {
	Env             string        `yaml:"env" env-required:"true"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env-required:"true"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-required:"true"`
	GRPC            GRPCConfig    `yaml:"grpc" env-required:"true"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
}

func MustConfig(log zerolog.Logger) *Config {
	configPath := "./config/app.yaml"

	cfg := &Config{}
	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to read config")
	}

	return cfg
}
