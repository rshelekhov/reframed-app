package model

type StatusName string

const (
	StatusNotStarted StatusName = "Not started"
	StatusPlanned    StatusName = "StatusPlanned"
	StatusCompleted  StatusName = "StatusCompleted"
	StatusArchived   StatusName = "StatusArchived"
)
