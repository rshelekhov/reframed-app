package model

import "time"

type (
	UserRequestData struct {
		Email           string `json:"email" validate:"required,email"`
		Password        string `json:"password" validate:"required,min=8"`
		UpdatedPassword string `json:"updated_password"`
	}

	UserResponseData struct {
		ID        string    `json:"id,omitempty"`
		Email     string    `json:"email,omitempty"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	ResetPasswordRequestData struct {
		Email string `json:"email" validate:"required,email"`
	}

	ChangePasswordRequestData struct {
		Password string `json:"password" validate:"required,min=8"`
	}
)
