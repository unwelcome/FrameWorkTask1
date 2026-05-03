package postgresDB

import (
	sharedPostgres "github.com/unwelcome/FrameWorkTask1/backend/shared/postgres"
)

type DatabaseRepository struct {
	User UserRepository
}

func NewDatabaseInstance(connectString string) *DatabaseRepository {
	db := sharedPostgres.Connect(connectString)

	sharedPostgres.Migrate(db, migrationQueries())

	return &DatabaseRepository{
		User: NewUserRepository(db),
	}
}
