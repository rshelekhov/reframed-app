package storage

import (
	"errors"
)

// Errors shouldn't depend on a particular storage implementation,
// so they are placed in the storage package
var (
	ErrUserNotFound = errors.New("user not found")
	ErrNoUsersFound = errors.New("no users found")

	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

const (
	UniqueConstraintViolation = "23505"
)
