package model

type StatusName string

const (
	StatusNotStarted StatusName = "Not started"
	StatusPlanned    StatusName = "Planned"
	StatusCompleted  StatusName = "Completed"
	StatusArchived   StatusName = "Archived"
)

func (s StatusName) String() string {
	return string(s)
}
