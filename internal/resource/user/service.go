package user

import (
	"github.com/google/uuid"
	"github.com/rshelekhov/remedi/internal/storage"
)

type Service struct {
	storage storage.UserStorage
}

func NewService(storage storage.UserStorage) *Service {
	return &Service{
		storage: storage,
	}
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
	id := uuid.New()
	return id, nil
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
