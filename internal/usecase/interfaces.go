// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
)

type (
	Usecase interface {
		User
	}

	// User defines the user use-cases
	User interface {
		CreateUser(ctx context.Context, user *model.User) (string, error)
		GetUser(ctx context.Context, id string) (model.GetUser, error)
		GetUsers(ctx context.Context, pgn model.Pagination) ([]*model.GetUser, error)
		UpdateUser(ctx context.Context, id string, user *model.UpdateUser) error
		DeleteUser(ctx context.Context, id string) error
	}

	// UserStorage defines the user repository
	UserStorage interface {
		CreateUser(ctx context.Context, user model.User) error
		GetUser(ctx context.Context, id string) (model.GetUser, error)
		GetUsers(ctx context.Context, pgn model.Pagination) ([]*model.GetUser, error)
		UpdateUser(ctx context.Context, user model.User) error
		DeleteUser(ctx context.Context, id string) error
	}
)
