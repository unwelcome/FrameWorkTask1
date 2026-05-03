package postgresDB

import (
	"database/sql"
	"strings"

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
	queries = append(queries,
		`DO $$ BEGIN
    	CREATE TYPE statuses AS ENUM (
				'open',
				'close'
			);
			EXCEPTION 
				WHEN duplicate_object THEN null;
		END $$;`,

		`DO $$ BEGIN
    	CREATE TYPE roles AS ENUM (
				'unemployed',
				'engineer',
				'manager',
				'analytic',
				'chief'
			);
			EXCEPTION 
				WHEN duplicate_object THEN null;
		END $$;`,

		`CREATE TABLE IF NOT EXISTS companies (
			uuid VARCHAR(36) UNIQUE NOT NULL,
			title varchar(255) NOT NULL,
			status statuses NOT NULL DEFAULT 'close',
			created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(36) NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS employees (
			company_uuid VARCHAR(36) NOT NULL,
			user_uuid VARCHAR(36) NOT NULL,
			role roles NOT NULL DEFAULT 'unemployed', 
			joined_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(company_uuid, user_uuid)
		);`,

		`ALTER TABLE employees 
			ADD CONSTRAINT fk_employees_company FOREIGN KEY (company_uuid) REFERENCES companies(uuid) ON DELETE CASCADE;`,
	)

	for _, query := range queries {
		_, err := m.db.Exec(query)
		if err != nil {
			// Игнорируем ошибку, если constraint уже существует
			if strings.Contains(err.Error(), "already exists") ||
				strings.Contains(err.Error(), "duplicate_object") ||
				strings.Contains(err.Error(), "constraint") {
				continue
			}
			log.Fatal().Err(err).Msg("migrator failed to execute query")
		}
	}

	log.Info().Msg("migration completed successfully")
}
