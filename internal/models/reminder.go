package models

import "time"

// Reminder DB models
type Reminder struct {
	ID        string     `db:"id" json:"id,omitempty"`
	Content   string     `db:"content" json:"content,omitempty"`
	Read      bool       `db:"read" json:"read,omitempty"`
	ActionID  string     `db:"action_id" json:"action_id,omitempty"`
	UserID    string     `db:"user_id" json:"user_id,omitempty"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
