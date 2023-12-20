package postgres

import "github.com/google/uuid"

// Reminder DB model
type Reminder struct {
	ID            uuid.UUID
	AppointmentID uuid.UUID
	UserID        uuid.UUID
	Content       string
	Read          bool
	CreatedAt     string
	UpdatedAt     string
}
