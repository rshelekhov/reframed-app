package model

import (
	"time"
)

// User DB model
type User struct {
	ID        string     `db:"id" json:"id" `
	Email     string     `db:"email" json:"email" validate:"required,email"`
	Password  string     `db:"password" json:"password" validate:"required,min=8"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type CreateUser struct {
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"min=8"`
}

// GetUser used in the response body and usecase layer for getting a user by ID
type GetUser struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateUser uses in the request body and usecase layer for updating a user by ID
type UpdateUser struct {
	Email    string `json:"email" db:"email" validate:"email"`
	Password string `json:"password" db:"password" validate:"min=8"`
}

type UserResponse struct {
	ID string `json:"id"`
}
