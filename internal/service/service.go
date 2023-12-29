package service

import (
	"github.com/rshelekhov/remedi/internal/model"
	"github.com/rshelekhov/remedi/internal/storage"
)

type service struct {
	storage storage.Storage
}

// New creates a new service (service) layer
func New(storage storage.Storage) Service {
	return &service{storage}
}

// Service is the common interface for all services
type Service interface {
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
