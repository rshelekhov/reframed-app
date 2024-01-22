package models

import "time"

// Attachment DB models
type Attachment struct {
	// TODO: add AttachmentType
	ID             string     `db:"id" json:"id,omitempty"`
	FileName       string     `db:"file_name" json:"file_name,omitempty"`
	FileURL        string     `db:"file_url" json:"file_url,omitempty"`
	AttachmentSize string     `db:"attachment_size" json:"attachment_size,omitempty"`
	GoalID         string     `db:"goal_id" json:"goal_id,omitempty"`
	UpdatedAt      *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt      *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
