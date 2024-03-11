package model

import "time"

// Reminder DB model
type Reminder struct {
	ID        string    `db:"id" json:"id,omitempty"`
	Content   string    `db:"content" json:"content,omitempty"`
	Read      bool      `db:"read" json:"read,omitempty"`
	TaskID    string    `db:"task_id" json:"task_id,omitempty"`
	UserID    string    `db:"user_id" json:"user_id,omitempty"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	DeletedAt time.Time `db:"deleted_at" json:"deleted_at"`
}
