package service

import (
	"github.com/google/uuid"
	"github.com/rshelekhov/remedi/internal/resource/user"
)

type UserService interface {
	ListUsers() ([]user.User, error)
	CreateUser(user user.CreateUser) (uuid.UUID, error)
	ReadUser(id uuid.UUID) (user.User, error)
	UpdateUser(id uuid.UUID) (user.User, error)
	DeleteUser(id uuid.UUID) error
}
