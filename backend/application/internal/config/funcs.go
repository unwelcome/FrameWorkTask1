package config

import (
	"fmt"
)

// GetDBConnectionString Строка подключения к Postgres для application сервиса
func (config *Config) GetDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Db.Host, config.Db.Port, config.ApplicationService.DBUser, config.ApplicationService.DBPassword, config.ApplicationService.DBName)
}
