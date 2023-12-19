package models

import (
	"github.com/google/uuid"
	"time"
)

type Appointment struct {
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
