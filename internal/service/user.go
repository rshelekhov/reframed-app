package service

import (
	"fmt"
	"github.com/rshelekhov/remedi/internal/model"
	"github.com/segmentio/ksuid"
	"time"
)

// CreateUser creates a new user
func (s *service) CreateUser(user *model.CreateUser) (string, error) {
	const op = "user.service.CreateUser"

	id := ksuid.New()

	entity := model.User{
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
		return "", fmt.Errorf("%s: failed to create user: %w", op, err)
	}

	return entity.ID, nil
}

// GetUser returns a user by ID
func (s *service) GetUser(id string) (model.GetUser, error) {
	// const op = "user.service.GetUser"
	return s.storage.GetUser(id)
}

// GetUsers returns a list of users
func (s *service) GetUsers(pgn model.Pagination) ([]model.GetUser, error) {
	// const op = "user.service.GetUsers"
	return s.storage.GetUsers(pgn)
}

// UpdateUser updates a user by ID
func (s *service) UpdateUser(id string, user *model.UpdateUser) error {
	const op = "user.service.UpdateUser"

	entity := model.User{
		ID:        id,
		Email:     user.Email,
		Password:  user.Password,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		UpdatedAt: time.Now().UTC(),
	}

	err := s.storage.UpdateUser(entity)
	if err != nil {
		return fmt.Errorf("%s: failed to update user: %w", op, err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (s *service) DeleteUser(id string) error {
	// const op = "user.service.DeleteUser"
	return s.storage.DeleteUser(id)
}

// GetUserRoles returns a list of roles
func (s *service) GetUserRoles() ([]model.GetRole, error) {
	// const op = "user.service.GetUserRoles"
	return s.storage.GetUserRoles()
}
