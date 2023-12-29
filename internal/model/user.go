package model

import (
	"time"
)

// User DB model
type User struct {
	ID        string     `db:"id"`
	Email     string     `db:"email"`
	Password  string     `db:"password"`
	RoleID    int        `db:"role_id"`
	FirstName string     `db:"first_name"`
	LastName  string     `db:"last_name"`
	Phone     string     `db:"phone"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// CreateUser uses in the request body and service layer for create a new user
type CreateUser struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	RoleID    int    `json:"role_id" validate:"required"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Phone     string `json:"phone" validate:"required,e164"`
}

// GetUser used in the response body and service layer for getting a user by ID
type GetUser struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	RoleID    int       `json:"role_id" db:"role_id"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Phone     string    `json:"phone" db:"phone"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// UpdateUser uses in the request body and service layer for updating a user by ID
type UpdateUser struct {
	Email     string    `json:"email" db:"email" validate:"email"`
	Password  string    `json:"password" db:"password" validate:"min=8"`
	FirstName string    `json:"first_name" db:"first_name"`
	LastName  string    `json:"last_name" db:"last_name"`
	Phone     string    `json:"phone" db:"phone" validate:"e164"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GetRole used in the response body and service layer
type GetRole struct {
	RoleID int    `json:"role_id" db:"id"`
	Title  string `json:"title" db:"title"`
}
