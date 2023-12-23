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

func (s Storage) ListUsers() ([]User, error) {
	const op = "user.storage.ListUsers"
	users := make([]User, 0)
	return users, nil
}

func (s Storage) CreateUser(user CreateUser) (uuid.UUID, error) {
	const op = "user.storage.CreateUser"

	var lastInsertID uuid.UUID
	// sqlStatement := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`

	return lastInsertID, nil
}

func (s Storage) ReadUser(id uuid.UUID) (User, error) {
	const op = "user.storage.GetUser"

	var user User

	return user, nil
}

func (s Storage) UpdateUser(id uuid.UUID) (User, error) {
	const op = "user.storage.UpdateUser"

	var user User

	return user, nil
}

func (s Storage) DeleteUser(id uuid.UUID) error {
	const op = "user.storage.DeleteUser"

	return nil
}
