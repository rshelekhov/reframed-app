package model

import "time"

type (
	UserRequestData struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	UserResponseData struct {
		ID        string    `json:"id,omitempty"`
		Email     string    `json:"email,omitempty"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	UpdateUserRequestData struct {
		Email           string `json:"email"`
		Password        string `json:"password"`
		UpdatedPassword string `json:"updated_password"`
	}

	ResetPasswordRequestData struct {
		Email string `json:"email" validate:"required,email"`
	}

	ChangePasswordRequestData struct {
		Password string `json:"password" validate:"required,min=8"`
	}
)
