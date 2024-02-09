package models

import "time"

// Pagination represents pagination parameters
type Pagination struct {
	Limit     int       `json:"limit"`
	AfterID   string    `json:"after_id"`
	AfterDate time.Time `json:"after_date"`
}
