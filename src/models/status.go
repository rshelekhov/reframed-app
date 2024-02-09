package models

// Status DB model
type Status struct {
	ID         string `db:"id" json:"id,omitempty"`
	StatusName string `db:"status_name" json:"status_name,omitempty"`
}
