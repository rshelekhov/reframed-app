package model

import "time"

// Pagination represents pagination parameters
type Pagination struct {
	Limit     int32     `json:"limit"`
	AfterID   string    `json:"after_id"`
	AfterDate time.Time `json:"after_date"`
}
