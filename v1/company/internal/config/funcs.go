package config

import "fmt"

// GetDBConnectionString Строка подключения к Postgres для company сервиса
func (config *Config) GetDBConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Db.Host, config.Db.Port, config.CompanyService.DBUser, config.CompanyService.DBPassword, config.CompanyService.DBName)
}
