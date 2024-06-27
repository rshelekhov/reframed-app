package model

import (
	"time"
)

// Heading DB model
type (
	Heading struct {
		ID        string    `db:"id"`
		Title     string    `db:"title"`
		ListID    string    `db:"list_id"`
		UserID    string    `db:"user_id"`
		IsDefault bool      `db:"is_default"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
		DeletedAt time.Time `db:"deleted_at"`
	}

	HeadingRequestData struct {
		ID     string `json:"heading_id"`
		Title  string `json:"title" validate:"required"`
		ListID string `json:"list_id"`
		UserID string `json:"user_id"`
	}

	HeadingResponseData struct {
		ID        string    `json:"heading_id,omitempty"`
		Title     string    `json:"title,omitempty"`
		ListID    string    `json:"list_id,omitempty"`
		UserID    string    `json:"user_id,omitempty"`
		CreatedAt time.Time `json:"created_at,omitempty"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
	}
)

type headingTitle string

func (t headingTitle) String() string {
	return string(t)
}

const (
	DefaultHeading headingTitle = "Default"
)
