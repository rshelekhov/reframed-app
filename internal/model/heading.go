package model

import (
	"time"
)

// Heading DB model
type (
	Heading struct {
		ID        string    `db:"id"`
		Title     string    `db:"headingTitle"`
		ListID    string    `db:"list_id"`
		UserID    string    `db:"user_id"`
		IsDefault bool      `db:"is_default"`
		UpdatedAt time.Time `db:"updated_at"`
		DeletedAt time.Time `db:"deleted_at"`
	}

	HeadingRequestData struct {
		ID     string `json:"id"`
		Title  string `json:"title" validate:"required"`
		ListID string `json:"list_id"`
		UserID string `json:"user_id"`
	}

	HeadingResponseData struct {
		ID        string    `json:"id,omitempty"`
		Title     string    `json:"title,omitempty"`
		ListID    string    `json:"list_id,omitempty"`
		UserID    string    `json:"user_id,omitempty"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

type headingTitle string

func (t headingTitle) String() string {
	return string(t)
}

const (
	DefaultHeading headingTitle = "Default"
)
