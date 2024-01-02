package service

import (
	"fmt"
	"github.com/rshelekhov/remedi/internal/model"
	"github.com/segmentio/ksuid"
	"time"
)

// CreateUser creates a new user
func (a *app) CreateUser(user *model.CreateUser) (string, error) {
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

	err := a.storage.CreateUser(entity)
	if err != nil {
		return "", fmt.Errorf("%s: failed to create user: %w", op, err)
	}

	return entity.ID, nil
}

// GetUser returns a user by ID
func (a *app) GetUser(id string) (model.GetUser, error) {
	// const op = "user.service.GetUser"
	return a.storage.GetUser(id)
}

// GetUsers returns a list of users
func (a *app) GetUsers(pgn model.Pagination) ([]model.GetUser, error) {
	// const op = "user.service.GetUsers"
	return a.storage.GetUsers(pgn)
}

// UpdateUser updates a user by ID
func (a *app) UpdateUser(id string, user *model.UpdateUser) error {
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

	err := a.storage.UpdateUser(entity)
	if err != nil {
		return fmt.Errorf("%s: failed to update user: %w", op, err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (a *app) DeleteUser(id string) error {
	// const op = "user.service.DeleteUser"
	return a.storage.DeleteUser(id)
}

// GetUserRoles returns a list of roles
func (a *app) GetUserRoles() ([]model.GetRole, error) {
	// const op = "user.service.GetUserRoles"
	return a.storage.GetUserRoles()
}
