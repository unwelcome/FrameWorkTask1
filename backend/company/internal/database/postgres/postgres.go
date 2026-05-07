package postgresDB

import (
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog/log"
	sharedPostgres "github.com/unwelcome/FrameWorkTask1/backend/shared/postgres"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type DatabaseRepository struct {
	Company CompanyRepository
}

func NewDatabaseInstance(connectString string) *DatabaseRepository {
	db := sharedPostgres.Connect(connectString)

	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load migration files")
	}

	driver, err := migratePostgres.WithInstance(db, &migratePostgres.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create migration driver")
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init migrator")
	}

	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal().Err(err).Msg("failed to apply migrations")
	}

	log.Info().Msg("migrations applied successfully")

	return &DatabaseRepository{
		Company: NewCompanyRepository(db),
	}
}
