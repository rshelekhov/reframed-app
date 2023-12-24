package user

import (
	"github.com/go-playground/validator"
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
	validator *validator.Validate
	storage   Storage
}

// NewService creates a new service
func NewService(validate *validator.Validate, storage Storage) Service {
	return &userService{validate, storage}
}

// ListUsers returns a list of users
func (s *userService) ListUsers() ([]User, error) {
	const op = "user.service.ListUsers"
	users, err := s.storage.ListUsers()
	if err != nil {
		return nil, err
	}
	return users, nil
}

// CreateUser creates a new user
func (s *userService) CreateUser(user CreateUser) (uuid.UUID, error) {
	const op = "user.service.CreateUser"
	id := uuid.New()
	return id, nil
}

// ReadUser returns a user by id
func (s *userService) ReadUser(id uuid.UUID) (User, error) {
	const op = "user.service.ReadUser"
	return s.storage.ReadUser(id)
}

// UpdateUser updates a user by id
func (s *userService) UpdateUser(id uuid.UUID) (User, error) {
	const op = "user.service.UpdateUser"
	return s.storage.UpdateUser(id)
}

// DeleteUser deletes a user by id
func (s *userService) DeleteUser(id uuid.UUID) error {
	const op = "user.service.DeleteUser"
	return s.storage.DeleteUser(id)
}
