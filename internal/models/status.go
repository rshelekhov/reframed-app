package models

// StatusGoal DB model
type StatusGoal struct {
	ID         string `db:"id" json:"id,omitempty"`
	StatusName string `db:"status_name" json:"status_name,omitempty"`
}

// StatusAction DB model
type StatusAction struct {
	ID         string `db:"id" json:"id,omitempty"`
	StatusName string `db:"status_name" json:"status_name,omitempty"`
}
