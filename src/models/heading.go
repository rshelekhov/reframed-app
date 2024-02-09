package models

import "time"

// Heading DB model
type (
	Heading struct {
		ID        string     `db:"id" json:"id,omitempty"`
		Title     string     `db:"title" json:"title,omitempty"`
		ListID    string     `db:"list_id" json:"list_id,omitempty"`
		UserID    string     `db:"user_id" json:"user_id,omitempty"`
		IsDefault bool       `db:"is_default" json:"is_default,omitempty"`
		UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
		DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	}

	UpdateHeading struct {
		Title     string     `db:"title" json:"title,omitempty"`
		ListID    string     `db:"list_id" json:"list_id,omitempty"`
		UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	}
)
