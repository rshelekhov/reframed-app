package model

import (
	"time"
)

// Task DB model
type (
	Task struct {
		ID          string    `db:"id"`
		Title       string    `db:"title"`
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
		CreatedAt   time.Time `db:"created_at"`
		UpdatedAt   time.Time `db:"updated_at"`
		DeletedAt   time.Time `db:"deleted_at"`
	}

	TaskRequestData struct {
		ID          string `json:"task_id"`
		Title       string `json:"title" validate:"required"`
		Description string `json:"description"`
		StartDate   string `json:"start_date"`
		Deadline    string `json:"deadline"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`

		StartDateParsed time.Time
		DeadlineParsed  time.Time
		StartTimeParsed time.Time
		EndTimeParsed   time.Time

		StatusID  int      `json:"status_id"`
		ListID    string   `json:"list_id"`
		HeadingID string   `json:"heading_id"`
		UserID    string   `json:"user_id"`
		Tags      []string `json:"tags"`
	}

	TaskResponseData struct {
		ID          string    `json:"task_id,omitempty"`
		Title       string    `json:"title,omitempty"`
		Description string    `json:"description,omitempty"`
		StartDate   time.Time `json:"start_date,omitempty"`
		Deadline    time.Time `json:"deadline,omitempty"`
		StartTime   time.Time `json:"start_time,omitempty"`
		EndTime     time.Time `json:"end_time,omitempty"`
		StatusID    int       `json:"status_id,omitempty"`
		ListID      string    `json:"list_id,omitempty"`
		HeadingID   string    `json:"heading_id,omitempty"`
		UserID      string    `json:"user_id,omitempty"`
		Tags        []string  `json:"tags,omitempty"`
		Overdue     bool      `json:"overdue,omitempty"`
		CreatedAt   time.Time `json:"created_at,omitempty"`
		UpdatedAt   time.Time `json:"updated_at,omitempty"`
	}

	TaskRequestTimeData struct {
		ID        string `json:"task_id"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`

		StartTimeParsed time.Time
		EndTimeParsed   time.Time

		UserID string `json:"user_id"`
	}

	TaskResponseTimeData struct {
		ID        string    `json:"task_id,omitempty"`
		StartTime time.Time `json:"start_time,omitempty"`
		EndTime   time.Time `json:"end_time,omitempty"`
		UserID    string    `json:"user_id,omitempty"`
		UpdatedAt time.Time `json:"updated_at,omitempty"`
	}

	TaskGroupRaw struct {
		StartDate time.Time `json:"start_date,omitempty"`
		Month     time.Time `json:"month,omitempty"`
		ListID    string    `json:"list_id,omitempty"`
		HeadingID string    `json:"heading_id,omitempty"`
		Tasks     []byte    `json:"tasks,omitempty"`
	}

	TodayTaskGroup struct {
		ListID string             `json:"list_id,omitempty"`
		Tasks  []TaskResponseData `json:"tasks,omitempty"`
	}

	UpcomingTaskGroup struct {
		StartDate time.Time          `json:"start_date,omitempty"`
		Tasks     []TaskResponseData `json:"tasks,omitempty"`
	}

	OverdueTaskGroup struct {
		ListID string             `json:"list_id,omitempty"`
		Tasks  []TaskResponseData `json:"tasks,omitempty"`
	}

	TaskGroupForSomeday struct {
		ListID string             `json:"list_id,omitempty"`
		Tasks  []TaskResponseData `json:"tasks,omitempty"`
	}

	TaskGroupWithHeading struct {
		HeadingID string             `json:"heading_id,omitempty"`
		Tasks     []TaskResponseData `json:"tasks,omitempty"`
	}

	CompletedTasksGroup struct {
		Month time.Time          `json:"month,omitempty"`
		Tasks []TaskResponseData `json:"tasks,omitempty"`
	}

	ArchivedTasksGroup struct {
		Month time.Time          `json:"month,omitempty"`
		Tasks []TaskResponseData `json:"tasks,omitempty"`
	}
)
