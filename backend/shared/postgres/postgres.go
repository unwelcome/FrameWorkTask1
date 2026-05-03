package postgres

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

func Connect(connectString string) *sql.DB {
	db, err := sql.Open("postgres", connectString)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}

	if err = db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("failed to ping postgres")
	}

	return db
}
