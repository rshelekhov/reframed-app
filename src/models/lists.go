package models

import "time"

// List DB model
type (
	List struct {
		ID        string     `db:"id" json:"id,omitempty"`
		Title     string     `db:"title" json:"title" validate:"required"`
		UserID    string     `db:"user_id" json:"user_id,omitempty"`
		UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
		DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	}

	UpdateList struct {
		Title string `db:"title" json:"title"`
	}
)
