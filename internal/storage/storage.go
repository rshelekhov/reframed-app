package storage

import (
	"context"
	"errors"
	"github.com/rshelekhov/reframed/internal/model"
)

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrNoUsersFound              = errors.New("no users found")
	ErrUserAlreadyExists         = errors.New("user with this email already exists")
	ErrEmailAlreadyTaken         = errors.New("this email already taken")
	ErrNoChangesDetected         = errors.New("no changes detected")
	ErrNoPasswordChangesDetected = errors.New("no password changes detected")
)

// UserStorage defines the user repository
type UserStorage interface {
	CreateUser(ctx context.Context, user model.User) error
	GetUserByID(ctx context.Context, id string) (model.User, error)
	GetUsers(ctx context.Context, pgn model.Pagination) ([]model.User, error)
	UpdateUser(ctx context.Context, user model.User) error
	DeleteUser(ctx context.Context, id string) error
}
