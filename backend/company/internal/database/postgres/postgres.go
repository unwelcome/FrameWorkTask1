package postgresDB

import (
	"database/sql"

	"github.com/rs/zerolog/log"
)

type DatabaseRepository struct {
	migrator *Migrator
	Company  CompanyRepository
}

func NewDatabaseInstance(connectString string) *DatabaseRepository {
	// Подключение к postgres
	db, err := sql.Open("postgres", connectString)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}

	// Проверка подключения
	err = db.Ping()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to ping postgres")
	}

	// Создаем структуру postgres репозиториев
	databaseRepository := &DatabaseRepository{}

	// Создаем репозитории
	databaseRepository.migrator = NewMigrator(db)
	databaseRepository.Company = NewCompanyRepository(db)

	// Запускаем миграцию
	databaseRepository.migrator.Migrate()

	return databaseRepository
}
