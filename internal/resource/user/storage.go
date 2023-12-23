package user

import (
	"database/sql"
	"github.com/google/uuid"
)

type Storage interface {
	ListUsers() ([]User, error)
	CreateUser(user CreateUser) (uuid.UUID, error)
	ReadUser(id uuid.UUID) (User, error)
	UpdateUser(id uuid.UUID) (User, error)
	DeleteUser(id uuid.UUID) error
}

type userStorage struct {
	db *sql.DB
}

func NewStorage(conn *sql.DB) Storage {
	return &userStorage{db: conn}
}

func (s *userStorage) ListUsers() ([]User, error) {
	const op = "user.storage.ListUsers"
	users := make([]User, 0)
	return users, nil
}

func (s *userStorage) CreateUser(user CreateUser) (uuid.UUID, error) {
	const op = "user.storage.CreateUser"

	var lastInsertID uuid.UUID
	// sqlStatement := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`

	return lastInsertID, nil
}

func (s *userStorage) ReadUser(id uuid.UUID) (User, error) {
	const op = "user.storage.ReadUser"

	var user User

	return user, nil
}

func (s *userStorage) UpdateUser(id uuid.UUID) (User, error) {
	const op = "user.storage.UpdateUser"

	var user User

	return user, nil
}

func (s *userStorage) DeleteUser(id uuid.UUID) error {
	const op = "user.storage.DeleteUser"

	return nil
}
