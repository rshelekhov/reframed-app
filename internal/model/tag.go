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
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
		DeletedAt time.Time `db:"deleted_at"`
	}

	TagRequestData struct {
		Title  string `json:"title" validate:"required"`
		UserID string `json:"user_id"`
	}

	TagResponseData struct {
		ID        string    `json:"tag_id,omitempty"`
		Title     string    `json:"title,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
	}
)
