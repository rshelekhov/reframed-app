package models

import "time"

// Goal DB models
type Goal struct {
	ID           string     `db:"id" json:"id,omitempty"`
	Title        string     `db:"title" json:"title,omitempty"`
	Description  string     `db:"description" json:"description,omitempty"`
	ResolutionID string     `db:"resolution_id" json:"resolution_id,omitempty"`
	Year         int        `db:"year" json:"year,omitempty"`
	Quarter      int        `db:"quarter" json:"quarter,omitempty"`
	Month        int        `db:"month" json:"month,omitempty"`
	Week         int        `db:"week" json:"week,omitempty"`
	StatusID     int        `db:"status_id" json:"status_id,omitempty"`
	UserID       string     `db:"user_id" json:"user_id,omitempty"`
	UpdatedAt    *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
