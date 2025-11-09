package postgresDB

import (
	"database/sql"
	"github.com/rs/zerolog/log"
)

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Migrate() {
	var queries []string
	queries = append(queries, `CREATE TABLE IF NOT EXISTS users (
		uuid VARCHAR(36) UNIQUE NOT NULL,
		role varchar(15) NOT NULL DEFAULT 'unknown',
		email varchar(255) UNIQUE NOT NULL,
    	password_hash VARCHAR(255) NOT NULL,
		first_name varchar(50) NOT NULL,
		last_name varchar(50) NOT NULL,
		patronymic varchar(50),
		created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP);`,
	)

	for _, query := range queries {
		_, err := m.db.Exec(query)
		if err != nil {
			log.Fatal().Err(err).Msg("migrator failed to execute query")
		}
	}

	log.Info().Msg("migration completed successfully")
}
