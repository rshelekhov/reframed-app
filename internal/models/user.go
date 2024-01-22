package models

import (
	"time"
)

// User DB models
type User struct {
	ID        string     `db:"id" json:"id,omitempty"`
	Email     string     `db:"email" json:"email,omitempty" validate:"required,email"`
	Password  string     `db:"password" json:"password,omitempty" validate:"required,min=8"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// UpdateUser uses in the request body and usecase layer for updating a user by ID
type UpdateUser struct {
	Email    string `json:"email" db:"email" validate:"omitempty,email"`
	Password string `json:"password" db:"password" validate:"omitempty,min=8"`
}
