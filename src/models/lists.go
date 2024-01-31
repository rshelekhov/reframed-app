package models

import "time"

// List DB models
type List struct {
	ID        string     `db:"id" json:"id,omitempty"`
	Title     string     `db:"title" json:"title,omitempty"`
	UserID    string     `db:"user_id" json:"user_id,omitempty"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

type UpdateList struct {
	Title string `db:"title" json:"title"`
}
