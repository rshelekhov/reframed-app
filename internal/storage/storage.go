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

type (
	// UserStorage defines the user repository
	UserStorage interface {
		CreateUser(ctx context.Context, user models.User) error
		GetUserByID(ctx context.Context, id string) (models.User, error)
		GetUsers(ctx context.Context, pgn models.Pagination) ([]models.User, error)
		UpdateUser(ctx context.Context, user models.User) error
		DeleteUser(ctx context.Context, id string) error
	}

	// ListStorage defines the list repository
	ListStorage interface {
		CreateList(ctx context.Context, list models.List) error
		GetListByID(ctx context.Context, id string) (models.List, error)
		GetLists(ctx context.Context, pgn models.Pagination) ([]models.List, error)
		UpdateList(ctx context.Context, list models.List) error
		DeleteList(ctx context.Context, id string) error
	}
)
