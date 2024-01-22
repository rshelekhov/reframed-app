package models

import "time"

// Resolution DB models
type Resolution struct {
	ID          string     `db:"id" json:"id,omitempty"`
	Title       string     `db:"title" json:"title,omitempty"`
	Description string     `db:"description" json:"description,omitempty"`
	Year        int        `db:"year" json:"year,omitempty"`
	UserID      string     `db:"user_id" json:"user_id,omitempty"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
