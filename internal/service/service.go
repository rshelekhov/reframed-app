package service

import (
	"errors"
	"github.com/go-playground/validator"
	"github.com/rshelekhov/remedi/internal/model"
	"github.com/rshelekhov/remedi/internal/storage"
)

var ErrValidationError = errors.New("validation error")

type app struct {
	storage  storage.Storage
	validate *validator.Validate
}

// New creates a new service layer
func New(storage storage.Storage, v *validator.Validate) App {
	return &app{storage, v}
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
