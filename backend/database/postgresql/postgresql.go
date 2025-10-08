package postgresdb

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
)

func Connect(connectString string, l zerolog.Logger) *sql.DB {
	postgres, err := ConnectToPostgres(connectString)
	if err != nil {
		l.Fatal().Err(err).Msg("Database connection failed")
	}

	l.Trace().Msg("Successfully connected to PostgresSQL!")
	return postgres
}

func ConnectToPostgres(connectString string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connectString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
