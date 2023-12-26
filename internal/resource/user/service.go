package user

import (
	"fmt"
	"github.com/segmentio/ksuid"
	"time"
)

type Service interface {
	ListUsers() ([]User, error)
	CreateUser(user CreateUser) (string, error)
	ReadUser(id string) (User, error)
	UpdateUser(id string) (string, error)
	DeleteUser(id string) error
}

type userService struct {
	storage Storage
}

// NewService creates a new service
func NewService(storage Storage) Service {
	return &userService{storage}
}

// ListUsers returns a list of users
func (s *userService) ListUsers() ([]User, error) {
	const op = "user.service.ListUsers"
	users, err := s.storage.ListUsers()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return users, nil
}

// CreateUser creates a new user
func (s *userService) CreateUser(user CreateUser) (string, error) {
	const op = "user.service.CreateUser"

	id := ksuid.New()

	entity := User{
		ID:        id.String(),
		Email:     user.Email,
		Password:  user.Password,
		RoleID:    user.RoleID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		UpdatedAt: time.Now().UTC(),
	}

	err := s.storage.CreateUser(entity)
	if err != nil {
		// TODO learn how to return nil instead of uuid.Nil
		return "", fmt.Errorf("%s: failed to create user: %w", op, err)
	}

	return entity.ID, nil
}

// ReadUser returns a user by id
func (s *userService) ReadUser(id string) (User, error) {
	const op = "user.service.ReadUser"
	return s.storage.ReadUser(id)
}

// UpdateUser updates a user by id
func (s *userService) UpdateUser(id string) (string, error) {
	const op = "user.service.UpdateUser"

	err := s.storage.UpdateUser(id)
	if err != nil {
		return "", fmt.Errorf("%s: failed to update user: %w", op, err)
	}

	return id, nil
}

// DeleteUser deletes a user by id
func (s *userService) DeleteUser(id string) error {
	const op = "user.service.DeleteUser"
	return s.storage.DeleteUser(id)
}
