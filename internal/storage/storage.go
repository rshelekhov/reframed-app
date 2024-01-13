package storage

import "errors"

var (
	ErrUserNotFound              = errors.New("user not found")
	ErrNoUsersFound              = errors.New("no users found")
	ErrUserAlreadyExists         = errors.New("user with this email already exists")
	ErrEmailAlreadyTaken         = errors.New("this email already taken")
	ErrNoChangesDetected         = errors.New("no changes detected")
	ErrNoPasswordChangesDetected = errors.New("no password changes detected")
)
