package storage

import (
	"context"
	"errors"
	"github.com/rshelekhov/reframed/internal/models"
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
	CreateUser(ctx context.Context, user models.User) error
	GetUserByID(ctx context.Context, id string) (models.User, error)
	GetUsers(ctx context.Context, pgn models.Pagination) ([]models.User, error)
	UpdateUser(ctx context.Context, user models.User) error
	DeleteUser(ctx context.Context, id string) error
}
