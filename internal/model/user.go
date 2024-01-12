package model

import (
	"time"
)

// User DB model
type User struct {
	ID        string     `db:"id" json:"id" `
	Email     string     `db:"email" json:"email" validate:"required,email"`
	Password  string     `db:"password" json:"password" validate:"required,min=8"`
	FirstName string     `db:"first_name" json:"first_name" validate:"required"`
	LastName  string     `db:"last_name" json:"last_name" validate:"required"`
	Phone     string     `db:"phone" json:"phone" validate:"required,e164"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// GetUser used in the response body and usecase layer for getting a user by ID
type GetUser struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Phone     string    `json:"phone" db:"phone"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateUser uses in the request body and usecase layer for updating a user by ID
type UpdateUser struct {
	Email     string `json:"email" db:"email" validate:"email"`
	Password  string `json:"password" db:"password" validate:"min=8"`
	FirstName string `json:"first_name" db:"first_name"`
	LastName  string `json:"last_name" db:"last_name"`
	Phone     string `json:"phone" db:"phone" validate:"e164"`
}
