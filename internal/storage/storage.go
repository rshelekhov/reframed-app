package storage

import (
	"errors"
	"github.com/rshelekhov/remedi/internal/model"
)

// Errors shouldn't depend on a particular storage implementation,
// so they are placed in the storage package
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrNoUsersFound        = errors.New("no users found")
	ErrNoRolesFound        = errors.New("no roles found")
	ErrRoleNotFound        = errors.New("role not found")
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrFileNotFound        = errors.New("file not found")

	ErrUserAlreadyExists        = errors.New("user with this email already exists")
	ErrAppointmentAlreadyExists = errors.New("appointment exists")
)

const (
	UniqueConstraintViolation = "23505"
)

// Storage is the common interface for all storage implementations
type Storage interface {
	UserStorage
}

// UserStorage is the interface that wraps the basic CRUD operations for users
type UserStorage interface {
	CreateUser(user *model.User) error
	GetUser(id string) (model.GetUser, error)
	GetUsers(model.Pagination) ([]model.GetUser, error)
	UpdateUser(user *model.User) error
	DeleteUser(id string) error
	GetUserRoles() ([]model.GetRole, error)
}
