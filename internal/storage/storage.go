package storage

import "errors"

// Errors shouldn't depend on a particular storage implementation,
// so they are placed in the storage package
var (
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrPhysicianNotFound   = errors.New("physician not found")
	ErrClientNotFound      = errors.New("client not found")
	ErrAssistantNotFound   = errors.New("assistant not found")
	ErrFileNotFound        = errors.New("file not found")

	ErrAppointmentExists = errors.New("appointment exists")
)
