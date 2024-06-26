package model

import "time"

// Pagination represents pagination parameters
type Pagination struct {
	Limit      int32     `json:"limit"`
	Cursor     string    `json:"cursor"`
	CursorDate time.Time `json:"cursor_date"`
}
