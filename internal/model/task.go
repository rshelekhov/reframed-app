package model

import (
	"time"
)

// Task DB model
type (
	Task struct {
		ID          string    `db:"id"`
		Title       string    `db:"headingTitle"`
		Description string    `db:"description"`
		StartDate   time.Time `db:"start_date"`
		Deadline    time.Time `db:"deadline"`
		StartTime   time.Time `db:"start_time"`
		EndTime     time.Time `db:"end_time"`
		StatusID    int       `db:"status_id"`
		ListID      string    `db:"list_id"`
		HeadingID   string    `db:"heading_id"`
		UserID      string    `db:"user_id"`
		Tags        []string
		Overdue     bool
		UpdatedAt   time.Time `db:"updated_at"`
		DeletedAt   time.Time `db:"deleted_at"`
	}

	TaskRequestData struct {
		ID          string    `json:"id"`
		Title       string    `json:"title" validate:"required"`
		Description string    `json:"description"`
		StartDate   time.Time `json:"start_date"`
		Deadline    time.Time `json:"deadline"`
		StartTime   time.Time `json:"start_time"`
		EndTime     time.Time `json:"end_time"`
		StatusID    int       `json:"status_id"`
		ListID      string    `json:"list_id"`
		HeadingID   string    `json:"heading_id"`
		UserID      string    `json:"user_id"`
		Tags        []string  `json:"tags"`
	}

	TaskResponseData struct {
		ID          string    `json:"id,omitempty"`
		Title       string    `json:"title,omitempty"`
		Description string    `json:"description,omitempty"`
		StartDate   time.Time `json:"start_date"`
		Deadline    time.Time `json:"deadline"`
		StartTime   time.Time `json:"start_time"`
		EndTime     time.Time `json:"end_time"`
		StatusID    int       `json:"status_id,omitempty"`
		ListID      string    `json:"list_id,omitempty"`
		HeadingID   string    `json:"heading_id,omitempty"`
		UserID      string    `json:"user_id,omitempty"`
		Tags        []string  `json:"tags,omitempty"`
		Overdue     bool      `json:"overdue,omitempty"`
		UpdatedAt   time.Time `json:"updated_at"`
	}

	TaskRequestTimeData struct {
		ID        string    `json:"id"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
		UserID    string    `json:"user_id"`
	}

	TaskResponseTimeData struct {
		ID        string    `json:"id"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
		UserID    string    `json:"user_id"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	TaskGroup struct {
		StartDate time.Time          `json:"start_date"`
		Month     int32              `json:"month,omitempty"`
		ListID    string             `json:"list_id,omitempty"`
		HeadingID string             `json:"heading_id,omitempty"`
		Tasks     []TaskResponseData `json:"tasks"`
	}
)
