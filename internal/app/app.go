package app

import (
	"github.com/rshelekhov/remedi/internal/model"
	"github.com/rshelekhov/remedi/internal/storage"
)

type app struct {
	storage storage.Storage
}

// New creates a new app layer
func New(storage storage.Storage) App {
	return &app{storage}
}

// App is the common interface for all services
type App interface {
	UserService
}

// UserService defines the user use-cases
type UserService interface {
	CreateUser(user *model.CreateUser) (string, error)
	GetUser(id string) (model.GetUser, error)
	GetUsers(model.Pagination) ([]model.GetUser, error)
	UpdateUser(id string, user *model.UpdateUser) error
	DeleteUser(id string) error
	GetUserRoles() ([]model.GetRole, error)
}
