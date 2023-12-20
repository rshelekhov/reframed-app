package postgres

import (
	"github.com/google/uuid"
	"time"
)

// MedicalReport DB model
type MedicalReport struct {
	ID              uuid.UUID
	Diagnosis       string
	Recommendations string
	AppointmentID   uuid.UUID
	Attachments     []Attachment
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Attachment DB model
type Attachment struct {
	ID              uuid.UUID
	FileName        string
	URL             string
	AttachmentSize  string
	MedicalReportID uuid.UUID
	AttachedByID    uuid.UUID
	AttachedAt      time.Time
	UpdatedAt       time.Time
}
