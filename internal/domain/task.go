package domain

import (
	"context"
	"time"
)

// Task DB domain
type (
	Task struct {
		ID          string     `db:"id"`
		Title       string     `db:"headingTitle"`
		Description string     `db:"description"`
		StartDate   *time.Time `db:"start_date"`
		Deadline    *time.Time `db:"deadline"`
		StartTime   *time.Time `db:"start_time"`
		EndTime     *time.Time `db:"end_time"`
		StatusID    int        `db:"status_id"`
		ListID      string     `db:"list_id"`
		HeadingID   string     `db:"heading_id"`
		UserID      string     `db:"user_id"`
		Tags        []string
		Overdue     bool
		UpdatedAt   *time.Time `db:"updated_at"`
		DeletedAt   *time.Time `db:"deleted_at"`
	}

	TaskRequestData struct {
		ID          string     `json:"id"`
		Title       string     `json:"title" validate:"required"`
		Description string     `json:"description"`
		StartDate   *time.Time `json:"start_date"`
		Deadline    *time.Time `json:"deadline"`
		StartTime   *time.Time `json:"start_time"`
		EndTime     *time.Time `json:"end_time"`
		StatusID    int        `json:"status_id"`
		ListID      string     `json:"list_id"`
		HeadingID   string     `json:"heading_id"`
		UserID      string     `json:"user_id"`
		Tags        []string   `json:"tags"`
	}

	TaskResponseData struct {
		ID          string     `json:"id,omitempty"`
		Title       string     `json:"title,omitempty"`
		Description string     `json:"description,omitempty"`
		StartDate   *time.Time `json:"start_date,omitempty"`
		Deadline    *time.Time `json:"deadline,omitempty"`
		StartTime   *time.Time `json:"start_time,omitempty"`
		EndTime     *time.Time `json:"end_time,omitempty"`
		StatusID    int        `json:"status_id,omitempty"`
		ListID      string     `json:"list_id,omitempty"`
		HeadingID   string     `json:"heading_id,omitempty"`
		UserID      string     `json:"user_id,omitempty"`
		Tags        []string   `json:"tags,omitempty"`
		Overdue     bool       `json:"overdue,omitempty"`
		UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	}

	TaskGroup struct {
		StartDate *time.Time         `json:"start_date,omitempty"`
		Month     *time.Time         `json:"month,omitempty"`
		ListID    string             `json:"list_id,omitempty"`
		HeadingID string             `json:"heading_id,omitempty"`
		Tasks     []TaskResponseData `json:"tasks"`
	}
)

type (
	TaskUsecase interface {
		CreateTask(ctx context.Context, data *TaskRequestData) (string, error)
		GetTaskByID(ctx context.Context, data TaskRequestData) (TaskResponseData, error)
		GetTasksByUserID(ctx context.Context, userID string, pgn Pagination) ([]TaskResponseData, error)
		GetTasksByListID(ctx context.Context, data TaskRequestData) ([]TaskResponseData, error)
		GetTasksGroupedByHeadings(ctx context.Context, data TaskRequestData) ([]TaskGroup, error)
		GetTasksForToday(ctx context.Context, userID string) ([]TaskGroup, error)
		GetUpcomingTasks(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		GetOverdueTasks(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		GetTasksForSomeday(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		GetCompletedTasks(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		GetArchivedTasks(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		UpdateTask(ctx context.Context, data *TaskRequestData) error
		UpdateTaskTime(ctx context.Context, data *TaskRequestData) error
		MoveTaskToAnotherList(ctx context.Context, data TaskRequestData) error
		CompleteTask(ctx context.Context, data TaskRequestData) error
		ArchiveTask(ctx context.Context, data TaskRequestData) error
	}

	TaskStorage interface {
		CreateTask(ctx context.Context, task Task) error
		GetTaskStatusID(ctx context.Context, status StatusName) (int, error)
		GetTaskByID(ctx context.Context, taskID, userID string) (Task, error)
		GetTasksByUserID(ctx context.Context, userID string, pgn Pagination) ([]Task, error)
		GetTasksByListID(ctx context.Context, listID, userID string) ([]Task, error)
		GetTasksGroupedByHeadings(ctx context.Context, listID, userID string) ([]TaskGroup, error)
		GetTasksForToday(ctx context.Context, userID string) ([]TaskGroup, error)
		GetUpcomingTasks(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		GetOverdueTasks(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		GetTasksForSomeday(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		GetCompletedTasks(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		GetArchivedTasks(ctx context.Context, userID string, pgn Pagination) ([]TaskGroup, error)
		UpdateTask(ctx context.Context, task Task) error
		UpdateTaskTime(ctx context.Context, task Task) error
		MoveTaskToAnotherList(ctx context.Context, task Task) error
		CompleteTask(ctx context.Context, task Task) error
		ArchiveTask(ctx context.Context, task Task) error
	}
)
