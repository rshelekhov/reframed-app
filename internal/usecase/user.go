package usecase

import (
	"fmt"
	"github.com/rshelekhov/reframed/internal/entity"
	"github.com/segmentio/ksuid"
	"time"
)

type UserUsecase struct {
	storage UserStorage
}

func NewUserUsecase(s UserStorage) *UserUsecase {
	return &UserUsecase{s}
}

// CreateUser creates a new user
func (uc *UserUsecase) CreateUser(user *entity.CreateUser) (string, error) {
	const op = "user.usecase.CreateUser"

	id := ksuid.New()

	entity := entity.User{
		ID:        id.String(),
		Email:     user.Email,
		Password:  user.Password,
		RoleID:    user.RoleID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		UpdatedAt: time.Now().UTC(),
	}

	err := uc.storage.CreateUser(&entity)
	if err != nil {
		return "", fmt.Errorf("%s: failed to create user: %w", op, err)
	}

	return entity.ID, nil
}

// GetUser returns a user by ID
func (uc *UserUsecase) GetUser(id string) (entity.GetUser, error) {
	// const op = "user.usecase.GetUser"
	return uc.storage.GetUser(id)
}

// GetUsers returns a list of users
func (uc *UserUsecase) GetUsers(pgn entity.Pagination) ([]entity.GetUser, error) {
	// const op = "user.usecase.GetUsers"
	return uc.storage.GetUsers(pgn)
}

// UpdateUser updates a user by ID
func (uc *UserUsecase) UpdateUser(id string, user *entity.UpdateUser) error {
	const op = "user.usecase.UpdateUser"

	entity := entity.User{
		ID:        id,
		Email:     user.Email,
		Password:  user.Password,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.Phone,
		UpdatedAt: time.Now().UTC(),
	}

	err := uc.storage.UpdateUser(&entity)
	if err != nil {
		return fmt.Errorf("%s: failed to update user: %w", op, err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (uc *UserUsecase) DeleteUser(id string) error {
	// const op = "user.usecase.DeleteUser"
	return uc.storage.DeleteUser(id)
}

// GetUserRoles returns a list of roles
func (uc *UserUsecase) GetUserRoles() ([]entity.GetRole, error) {
	// const op = "user.usecase.GetUserRoles"
	return uc.storage.GetUserRoles()
}
