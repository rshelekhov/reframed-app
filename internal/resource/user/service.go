package user

import (
	"github.com/google/uuid"
)

type Service struct {
	storage *Storage
}

func (s *Service) ListUsers() ([]User, error) {
	const op = "user.service.ListUsers"
	users, err := s.storage.ListUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Service) CreateUser(user CreateUser) (uuid.UUID, error) {
	return s.storage.CreateUser(user)
}

func (s *Service) ReadUser(id uuid.UUID) (User, error) {
	return s.storage.ReadUser(id)
}

func (s *Service) UpdateUser(id uuid.UUID) (User, error) {
	return s.storage.UpdateUser(id)
}

func (s *Service) DeleteUser(id uuid.UUID) error {
	return s.storage.DeleteUser(id)
}
