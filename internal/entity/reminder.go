package entity

import "github.com/google/uuid"

// Reminder DB entity
type Reminder struct {
	ID            uuid.UUID // TODO: change to string (ksuid)
	AppointmentID uuid.UUID
	UserID        uuid.UUID
	Content       string
	Read          bool
	CreatedAt     string
	UpdatedAt     string
}
