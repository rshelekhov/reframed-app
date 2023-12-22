package user

import (
	"database/sql"
	"github.com/google/uuid"
)

type Storage struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Storage {
	return &Storage{
		db: db,
	}
}

func (s Storage) CreateUser(user User) (uuid.UUID, error) {
	const op = "storage.postgres.CreateUser"

	var lastInsertID uuid.UUID
	// sqlStatement := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`

	return lastInsertID, nil
}

func (s Storage) ReadUser(id uuid.UUID) (User, error) {
	const op = "storage.postgres.GetUser"

	var user User

	return user, nil
}

func (s Storage) UpdateUser(user User) error {
	const op = "storage.postgres.UpdateUser"
	return nil
}

func (s Storage) DeleteUser(id uuid.UUID) error {
	const op = "storage.postgres.DeleteUser"
	return nil
}
