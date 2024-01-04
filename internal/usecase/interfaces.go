// Package usecase implements application business logic. Each logic group in own file.
package usecase

import "github.com/rshelekhov/reframed/internal/entity"

type (
	Usecase interface {
		User
	}

	// User defines the user use-cases
	User interface {
		CreateUser(user *entity.CreateUser) (string, error)
		GetUser(id string) (entity.GetUser, error)
		GetUsers(pgn entity.Pagination) ([]entity.GetUser, error)
		UpdateUser(id string, user *entity.UpdateUser) error
		DeleteUser(id string) error
		GetUserRoles() ([]entity.GetRole, error)
	}

	// UserStorage defines the user repository
	UserStorage interface {
		CreateUser(user *entity.User) error
		GetUser(id string) (entity.GetUser, error)
		GetUsers(pgn entity.Pagination) ([]entity.GetUser, error)
		UpdateUser(user *entity.User) error
		DeleteUser(id string) error
		GetUserRoles() ([]entity.GetRole, error)
	}
)
