package models

// Priority DB models
type Priority struct {
	ID           string `db:"id" json:"id,omitempty"`
	PriorityName string `db:"priority_name" json:"priority_name,omitempty"`
}
