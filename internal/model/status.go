package model

type StatusName string

const (
	StatusNotStarted StatusName = "Not started"
	StatusPlanned    StatusName = "StatusPlanned"
	StatusCompleted  StatusName = "StatusCompleted"
	StatusArchived   StatusName = "StatusArchived"
)

func (s StatusName) String() string {
	return string(s)
}
