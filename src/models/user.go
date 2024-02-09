package models

import (
	"time"
)

// User DB models
type (
	UserRequestData struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	User struct {
		ID           string     `db:"id" json:"id,omitempty"`
		Email        string     `db:"email" json:"email,omitempty"`
		PasswordHash string     `db:"password_hash" json:"password_hash,omitempty"`
		UpdatedAt    *time.Time `db:"updated_at" json:"updated_at,omitempty"`
		DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	}
)
