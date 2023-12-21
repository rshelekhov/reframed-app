package models

import (
	"github.com/google/uuid"
	"time"
)

// Appointment DB model
type AppointmentStorage struct {
	ID               uuid.UUID
	DoctorID         uuid.UUID
	ClientID         uuid.UUID
	Title            string
	Content          string
	StatusID         int
	ScheduledAt      time.Time
	FirstAppointment bool
	CreatedByID      uuid.UUID
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
