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
