package postgres

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

const (
	maxRetries = 10
	retryDelay = 3 * time.Second
)

func Connect(connectString string) *sql.DB {
	db, err := sql.Open("postgres", connectString)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open postgres connection")
	}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Warn().
			Err(err).
			Int("attempt", attempt).
			Int("max", maxRetries).
			Msgf("failed to ping postgres, retrying in %s...", retryDelay)
		time.Sleep(retryDelay)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("failed to ping postgres after all retries")
	}

	return db
}
