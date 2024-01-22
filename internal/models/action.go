package models

import "time"

// Action DB models
type Action struct {
	ID          string     `db:"id" json:"id,omitempty"`
	Title       string     `db:"title" json:"title,omitempty"`
	Description string     `db:"description" json:"description,omitempty"`
	GoalID      string     `db:"goal_id" json:"goal_id,omitempty"`
	StartDate   *time.Time `db:"start_date" json:"start_date,omitempty"`
	Deadline    *time.Time `db:"deadline" json:"deadline,omitempty"`
	StartTime   *time.Time `db:"start_time" json:"start_time,omitempty"`
	EndTime     *time.Time `db:"end_time" json:"end_time,omitempty"`
	PriorityID  int        `db:"priority_id" json:"priority_id,omitempty"`
	StatusID    int        `db:"status_id" json:"status_id,omitempty"`
	UserID      string     `db:"user_id" json:"user_id,omitempty"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at,omitempty"`
	DeletedAt   *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}
