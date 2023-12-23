package user

import (
	"github.com/google/uuid"
)

type Service interface {
	ListUsers() ([]User, error)
	CreateUser(user CreateUser) (uuid.UUID, error)
	ReadUser(id uuid.UUID) (User, error)
	UpdateUser(id uuid.UUID) (User, error)
	DeleteUser(id uuid.UUID) error
}

type userService struct {
	storage Storage
}

func NewService(repo Storage) Service {
	return &userService{repo}
}

func (s *userService) ListUsers() ([]User, error) {
	const op = "user.service.ListUsers"
	users, err := s.storage.ListUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *userService) CreateUser(user CreateUser) (uuid.UUID, error) {
	const op = "user.service.CreateUser"
	id := uuid.New()
	return id, nil
}

func (s *userService) ReadUser(id uuid.UUID) (User, error) {
	const op = "user.service.ReadUser"
	return s.storage.ReadUser(id)
}

func (s *userService) UpdateUser(id uuid.UUID) (User, error) {
	const op = "user.service.UpdateUser"
	return s.storage.UpdateUser(id)
}

func (s *userService) DeleteUser(id uuid.UUID) error {
	const op = "user.service.DeleteUser"
	return s.storage.DeleteUser(id)
}
