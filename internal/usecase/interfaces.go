// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"github.com/rshelekhov/reframed/internal/entity"
)

type (
	Usecase interface {
		User
	}

	// User defines the user use-cases
	User interface {
		CreateUser(ctx context.Context, user *entity.CreateUser) (string, error)
		GetUser(ctx context.Context, id string) (entity.GetUser, error)
		GetUsers(ctx context.Context, pgn entity.Pagination) ([]*entity.GetUser, error)
		UpdateUser(ctx context.Context, id string, user *entity.UpdateUser) error
		DeleteUser(ctx context.Context, id string) error
		GetUserRoles(ctx context.Context) ([]*entity.GetRole, error)
	}

	// UserStorage defines the user repository
	UserStorage interface {
		CreateUser(ctx context.Context, user entity.User) error
		GetUser(ctx context.Context, id string) (entity.GetUser, error)
		GetUsers(ctx context.Context, pgn entity.Pagination) ([]*entity.GetUser, error)
		UpdateUser(ctx context.Context, user entity.User) error
		DeleteUser(ctx context.Context, id string) error
		GetUserRoles(ctx context.Context) ([]*entity.GetRole, error)
	}
)
