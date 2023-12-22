package user

import (
	"github.com/google/uuid"
	"log/slog"
)

type Service struct {
	logger  *slog.Logger
	storage *Storage
}

func (s *Service) ListUsers() Users {
	return s.storage.ListUsers()
}

func (s *Service) CreateUser(user User) (uuid.UUID, error) {
	return s.storage.CreateUser(user)
}

func (s *Service) ReadUser(id uuid.UUID) (User, error) {
	return s.storage.ReadUser(id)
}

func (s *Service) UpdateUser(user User) error {
	return s.storage.UpdateUser(user)
}

func (s *Service) DeleteUser(id uuid.UUID) error {
	return s.storage.DeleteUser(id)
}
