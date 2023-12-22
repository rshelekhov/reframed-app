package postgres

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// TODO implement sqlx

type Storage struct {
	DB *sql.DB
}

// NewStorage ...
func NewStorage(dsn string) (*Storage, error) {
	const op = "storage.postgres.NewStorage"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db}, nil
}

// Close ...
func (s *Storage) Close() error {
	return s.DB.Close()
}
