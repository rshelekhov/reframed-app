package storage

import (
	"errors"
	"github.com/google/uuid"
	"github.com/rshelekhov/remedi/internal/resource/user"
)

// Errors shouldn't depend on a particular storage implementation,
// so they are placed in the storage package
var (
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrPhysicianNotFound   = errors.New("physician not found")
	ErrClientNotFound      = errors.New("client not found")
	ErrAssistantNotFound   = errors.New("assistant not found")
	ErrFileNotFound        = errors.New("file not found")

	ErrAppointmentExists = errors.New("appointment exists")
)

type UserStorage interface {
	ListUsers() ([]user.User, error)
	CreateUser(user user.User) (uuid.UUID, error)
	ReadUser(id uuid.UUID) (user.User, error)
	UpdateUser(id uuid.UUID) (user.User, error)
	DeleteUser(id uuid.UUID) error
}
