package postgresDB

import (
	sharedPostgres "github.com/unwelcome/FrameWorkTask1/backend/shared/postgres"
)

type DatabaseRepository struct {
	Company CompanyRepository
}

func NewDatabaseInstance(connectString string) *DatabaseRepository {
	db := sharedPostgres.Connect(connectString)

	sharedPostgres.Migrate(db, migrationQueries())

	return &DatabaseRepository{
		Company: NewCompanyRepository(db),
	}
}
