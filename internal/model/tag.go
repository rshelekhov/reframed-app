package model

import (
	"time"
)

// Tag DB model
type (
	Tag struct {
		ID        string    `db:"id" json:"id,omitempty"`
		Title     string    `db:"headingTitle" json:"headingTitle,omitempty"`
		UserID    string    `db:"user_id" json:"user_id,omitempty"`
		UpdatedAt time.Time `db:"updated_at" json:"updated_at,omitempty"`
		DeletedAt time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	}

	TagRequestData struct {
		Title  string `json:"title" validate:"required"`
		UserID string `json:"user_id"`
	}

	TagResponseData struct {
		ID        string    `json:"id,omitempty"`
		Title     string    `json:"title,omitempty"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
	}
)
