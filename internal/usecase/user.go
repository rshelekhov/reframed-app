package usecase

import (
	"context"
	"fmt"
	"github.com/rshelekhov/reframed/internal/model"
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
func (uc *UserUsecase) CreateUser(ctx context.Context, user *model.User) (string, error) {
	const op = "user.usecase.CreateUser"

	if user.ID == "" {
		user.ID = ksuid.New().String()
	}

	newUser := model.User{
		ID:        user.ID,
		Email:     user.Email,
		Password:  user.Password,
		UpdatedAt: time.Now().UTC(),
	}

	err := uc.storage.CreateUser(ctx, newUser)
	if err != nil {
		return "", fmt.Errorf("%s: failed to create user: %w", op, err)
	}

	return newUser.ID, nil
}

// GetUser returns a user by ID
func (uc *UserUsecase) GetUser(ctx context.Context, id string) (model.GetUser, error) {
	// const op = "user.usecase.GetUser"
	return uc.storage.GetUser(ctx, id)
}

// GetUsers returns a list of users
func (uc *UserUsecase) GetUsers(ctx context.Context, pgn model.Pagination) ([]*model.GetUser, error) {
	// const op = "user.usecase.GetUsers"
	return uc.storage.GetUsers(ctx, pgn)
}

// UpdateUser updates a user by ID
func (uc *UserUsecase) UpdateUser(ctx context.Context, id string, user *model.UpdateUser) error {
	const op = "user.usecase.UpdateUser"

	newUser := model.User{
		ID:        id,
		Email:     user.Email,
		Password:  user.Password,
		UpdatedAt: time.Now().UTC(),
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
