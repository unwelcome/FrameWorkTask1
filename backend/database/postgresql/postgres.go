package postgresql

import (
	"backend/internal/config"
	"database/sql"
	"fmt"
	"log"
)

type Database struct {
	DB *sql.DB
}

func Connect(cfg *config.Config) (*Database, error) {
	connStr := cfg.DBConnString()

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgresSQL!")
	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}
