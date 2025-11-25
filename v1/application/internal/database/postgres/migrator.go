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
			CREATE TYPE STATUSES AS ENUM (
				'created',
				'assigned',
				'in_procgress',
				'on_hold',
				'completed',
				'cancelled',
				'rejected',
				'failed',
				'archived',
			);
			EXCEPTION 
				WHEN duplicate_object THEN null;
		END $$;
		`,

		`CREATE TABLE IF NOT EXISTS applications (
			uuid VARCHAR(36) UNIQUE NOT NULL,
			version INTEGER DEFAULT 1,

			title VARCHAR(255) NOT NULL,
			description TEXT,
			status STATUSES NOT NULL,
			fix_log TEXT[] DEFAULT NULL,

			managed_by VARCHAR(36) DEFAULT NULL,
			executed_by VARCHAR(36) DEFAULT NULL,

			created_by VARCHAR(36) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			closed_at TIMESTAMP DEFAULT NULL,
			deleted_at TIMESTAMP DEFAULT NULL);
		`,
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
