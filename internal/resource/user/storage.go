package user

import (
	"database/sql"
	"github.com/google/uuid"
)

// TODO: implement sqlx.DB

type Storage interface {
	ListUsers() ([]User, error)
	CreateUser(user User) (uuid.UUID, error)
	ReadUser(id uuid.UUID) (User, error)
	UpdateUser(id uuid.UUID) (User, error)
	DeleteUser(id uuid.UUID) error
}

type userStorage struct {
	db *sql.DB
}

// NewStorage creates a new storage
func NewStorage(conn *sql.DB) Storage {
	return &userStorage{db: conn}
}

// ListUsers returns a list of users
func (s *userStorage) ListUsers() ([]User, error) {
	const op = "user.storage.ListUsers"
	users := make([]User, 0)
	return users, nil
}

// CreateUser creates a new user
func (s *userStorage) CreateUser(user User) (uuid.UUID, error) {
	const op = "user.storage.CreateUser"

	var lastInsertID uuid.UUID
	// sqlStatement := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`

	return lastInsertID, nil
}

// ReadUser returns a user by id
func (s *userStorage) ReadUser(id uuid.UUID) (User, error) {
	const op = "user.storage.ReadUser"

	var user User

	return user, nil
}

// UpdateUser updates a user by id
func (s *userStorage) UpdateUser(id uuid.UUID) (User, error) {
	const op = "user.storage.UpdateUser"

	var user User

	return user, nil
}

// DeleteUser deletes a user by id
func (s *userStorage) DeleteUser(id uuid.UUID) error {
	const op = "user.storage.DeleteUser"

	return nil
}
