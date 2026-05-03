package postgres

import (
	"database/sql"
	"strings"

	"github.com/rs/zerolog/log"
)

func Migrate(db *sql.DB, queries []string) {
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
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
