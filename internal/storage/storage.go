package storage

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrNoUsersFound = errors.New("no users found")

	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

const (
	UniqueConstraintViolation = "23505"
)
