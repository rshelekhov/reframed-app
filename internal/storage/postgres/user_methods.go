package postgres

import (
	"github.com/google/uuid"
	"github.com/rshelekhov/remedi/internal/models"
)

func (s *Storage) CreateUser(user models.User) (uuid.UUID, error) {
	const op = "storage.postgres.CreateUser"

	var lastInsertID uuid.UUID
	// sqlStatement := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`

	return lastInsertID, nil
}

func (s *Storage) GetUser(id uuid.UUID) (models.User, error) {
	const op = "storage.postgres.GetUser"

	var user models.User

	return user, nil
}

func (s *Storage) UpdateUser(user models.User) error {
	const op = "storage.postgres.UpdateUser"
	return nil
}

func (s *Storage) DeleteUser(id uuid.UUID) error {
	const op = "storage.postgres.DeleteUser"
	return nil
}
