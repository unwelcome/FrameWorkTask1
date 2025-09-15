package postgresql

import (
	"backend/internal/config"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
)

type Database struct {
	DB *sql.DB
}

func Connect(cfg *config.Config, l *slog.Logger) *Database {
	postgres, err := ConnectToPostgres(cfg)
	if err != nil {
		l.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	l.Info("Successfully connected to PostgresSQL!")
	return postgres
}

func ConnectToPostgres(cfg *config.Config) (*Database, error) {
	connStr := cfg.DBConnString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
