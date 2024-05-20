package model

import (
	"time"
)

// Tag DB model
type (
	Tag struct {
		ID        string    `db:"id"`
		Title     string    `db:"title"`
		UserID    string    `db:"user_id"`
		UpdatedAt time.Time `db:"updated_at"`
		DeletedAt time.Time `db:"deleted_at"`
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
