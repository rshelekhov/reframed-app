package usecase

import (
	"context"
	"fmt"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/storage"
	"github.com/segmentio/ksuid"
	"time"
)

type UserUsecase struct {
	storage storage.UserStorage
}

func NewUserUsecase(s storage.UserStorage) *UserUsecase {
	return &UserUsecase{s}
}

// CreateUser creates a new user
func (uc *UserUsecase) CreateUser(ctx context.Context, user *model.User) (string, error) {
	const op = "user.usecase.CreateUser"

	id := ksuid.New().String()
	now := time.Now().UTC()

	newUser := model.User{
		ID:        id,
		Email:     user.Email,
		Password:  user.Password,
		UpdatedAt: &now,
	}

	err := uc.storage.CreateUser(ctx, newUser)
	if err != nil {
		return "", fmt.Errorf("%s: failed to create user: %w", op, err)
	}

	return newUser.ID, nil
}

// GetUserByID returns a user by ID
func (uc *UserUsecase) GetUserByID(ctx context.Context, id string) (model.User, error) {
	// const op = "user.usecase.GetUserByID"
	return uc.storage.GetUserByID(ctx, id)
}

// GetUsers returns a list of users
func (uc *UserUsecase) GetUsers(ctx context.Context, pgn model.Pagination) ([]model.User, error) {
	// const op = "user.usecase.GetUsers"
	return uc.storage.GetUsers(ctx, pgn)
}

// UpdateUser updates a user by ID
func (uc *UserUsecase) UpdateUser(ctx context.Context, id string, user *model.UpdateUser) error {
	const op = "user.usecase.UpdateUser"
	now := time.Now().UTC()

	newUser := model.User{
		ID:        id,
		Email:     &user.Email,
		Password:  &user.Password,
		UpdatedAt: &now,
	}

	err := uc.storage.UpdateUser(ctx, newUser)
	if err != nil {
		return fmt.Errorf("%s: failed to update user: %w", op, err)
	}

	return nil
}

// DeleteUser deletes a user by ID
func (uc *UserUsecase) DeleteUser(ctx context.Context, id string) error {
	// const op = "user.usecase.DeleteUser"
	return uc.storage.DeleteUser(ctx, id)
}
