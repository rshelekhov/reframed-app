package postgres

import (
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

// TODO implement sqlx

type Storage struct {
	DB *sqlx.DB
}

// NewStorage ...
func NewStorage(dsn string) (*Storage, error) {
	const op = "storage.postgres.NewStorage"

	db, err := sqlx.Open("pgx", dsn)
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
