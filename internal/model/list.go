package model

import (
	"time"
)

// List DB model
type (
	List struct {
		ID        string    `db:"id"`
		Title     string    `db:"title"`
		UserID    string    `db:"user_id"`
		IsDefault bool      `db:"is_default"`
		UpdatedAt time.Time `db:"updated_at"`
		DeletedAt time.Time `db:"deleted_at"`
	}

	ListRequestData struct {
		ID     string `json:"list_id"`
		Title  string `json:"title" validate:"required"`
		UserID string `json:"user_id"`
	}

	ListResponseData struct {
		ID        string    `json:"list_id"`
		Title     string    `json:"title"`
		UserID    string    `json:"user_id"`
		UpdatedAt time.Time `json:"updated_at"`
	}
)

type listTitle string

func (t listTitle) String() string {
	return string(t)
}

const (
	DefaultInboxList listTitle = "Inbox"
)
