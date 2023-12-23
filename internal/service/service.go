package service

import (
	"github.com/google/uuid"
	"github.com/rshelekhov/remedi/internal/resource/user"
)

type UserService interface {
	CreateUserService(user user.CreateUser) (uuid.UUID, error)
}
