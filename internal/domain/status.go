package domain

// Status DB domain
type Status struct {
	ID         string `db:"id" json:"id,omitempty"`
	StatusName string `db:"status_name" json:"status_name,omitempty"`
}

type StatusName string

const (
	StatusNotStarted StatusName = "Not started"
	StatusPlanned    StatusName = "StatusPlanned"
	StatusCompleted  StatusName = "StatusCompleted"
	StatusArchived   StatusName = "StatusArchived"
)
