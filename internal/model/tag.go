package model

import (
	"time"
)

// Tag DB model
type (
	Tag struct {
		ID        string
		Title     string
		UserID    string
		UpdatedAt time.Time
		DeletedAt time.Time
	}

	TagRequestData struct {
		Title  string `json:"title" validate:"required"`
		UserID string `json:"user_id"`
	}

	TagResponseData struct {
		ID        string    `json:"id"`
		Title     string    `json:"title"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)
