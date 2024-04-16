package model

import (
	"time"
)

// User DB model
type (
	User struct {
		ID           string    `db:"id"`
		Email        string    `db:"email"`
		PasswordHash string    `db:"password_hash"`
		UpdatedAt    time.Time `db:"updated_at"`
		DeletedAt    time.Time `db:"deleted_at"`
	}

	UserRequestData struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
		AppID    int32  `json:"app_id" validate:"required"`
	}

	UserResponseData struct {
		ID        string    `json:"id,omitempty"`
		Email     string    `json:"email,omitempty"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)
