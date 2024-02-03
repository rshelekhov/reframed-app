package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/models"
	"strconv"
	"time"
)

type TaskStorage struct {
	*pgxpool.Pool
}

func NewTaskStorage(pool *pgxpool.Pool) *TaskStorage {
	return &TaskStorage{Pool: pool}
}

func (s *TaskStorage) CreateTask(ctx context.Context, task models.Task) error {
	const (
		op = "task.storage.CreateTask"

		querySelectStatus = `SELECT	id FROM statuses WHERE status_name = $1`

		queryInsertTask = `
			INSERT INTO tasks
    		(
				id,
				title,
				description,
				start_date,
				deadline,
				start_time,
				end_time,
				status_id,
				list_id,
				user_id,
				updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	)

	var status string
	err := s.QueryRow(ctx, querySelectStatus, c.StatusNotStarted).Scan(&status)
	if err != nil {
		return fmt.Errorf("%s: failed to get status: %w", op, err)
	}

	_, err = s.Exec(
		ctx,
		queryInsertTask,
		task.ID,
		task.Title,
		task.Description,
		task.StartDate,
		task.Deadline,
		task.StartTime,
		task.EndTime,
		task.StatusID,
		task.ListID,
		task.UserID,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("%s: failed to insert new task: %w", op, err)
	}

	return nil
}

func (s *TaskStorage) GetTaskByID(ctx context.Context, taskID, userID string) (models.Task, error) {
	const (
		op = "task.storage.GetTaskByID"

		query = `
			SELECT
				title,
				description,
				start_date,
				deadline,
				start_time,
				end_time,
				status_id,
				list_id,
				updated_at
			FROM tasks
			WHERE id = $1
			AND user_id = $2
			AND deleted_at IS NULL`
	)

	var task models.Task

	err := s.QueryRow(
		ctx,
		query,
		taskID,
		userID,
	).Scan(
		&task.Title,
		&task.Description,
		&task.StartDate,
		&task.Deadline,
		&task.StartTime,
		&task.EndTime,
		&task.StatusID,
		&task.ListID,
		&task.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return task, c.ErrTaskNotFound
	}
	if err != nil {
		return task, fmt.Errorf("%s: failed to get task: %w", op, err)
	}

	task.ID = taskID
	task.UserID = userID

	return task, nil
}

func (s *TaskStorage) GetTasksByUserID(ctx context.Context, userID string, pgn models.Pagination) ([]models.Task, error) {
	const (
		op = "task.storage.GetTasksByUserID"

		query = `
			SELECT
				id,
				title,
				description,
				start_date,
				deadline,
				start_time,
				end_time,
				status_id,
				list_id,
				updated_at
			FROM tasks
			WHERE user_id = $1
			AND deleted_at IS NULL
			ORDER BY id DESC LIMIT $2 OFFSET $3`
	)

	rows, err := s.Query(
		ctx,
		query,
		userID,
		pgn.Limit,
		pgn.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks: %w", op, err)
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		task := models.Task{}

		err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.StartDate,
			&task.Deadline,
			&task.StartTime,
			&task.EndTime,
			&task.StatusID,
			&task.ListID,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}
		task.UserID = userID
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(tasks) == 0 {
		return tasks, c.ErrNoTasksFound
	}

	return tasks, nil
}

func (s *TaskStorage) GetTasksByListID(ctx context.Context, listID, userID string, pgn models.Pagination) ([]models.Task, error) {
	const (
		op = "task.storage.GetTasksByListID"

		query = `
			SELECT
				id,
				title,
				description,
				start_date,
				deadline,
				start_time,
				end_time,
				status_id,
				updated_at
			FROM tasks
			WHERE list_id = $1
			AND user_id = $2
			AND deleted_at IS NULL
			ORDER BY id DESC LIMIT $3 OFFSET $4`
	)

	rows, err := s.Query(
		ctx,
		query,
		listID,
		userID,
		pgn.Limit,
		pgn.Offset,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks: %w", op, err)
	}
	defer rows.Close()

	var tasks []models.Task

	for rows.Next() {
		task := models.Task{}

		err = rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.StartDate,
			&task.Deadline,
			&task.StartTime,
			&task.EndTime,
			&task.StatusID,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}
		task.UserID = userID
		task.ListID = listID
		tasks = append(tasks, task)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(tasks) == 0 {
		return tasks, c.ErrNoTasksFound
	}

	return tasks, nil
}

func (s *TaskStorage) UpdateTask(ctx context.Context, task models.Task) error {
	const op = "task.storage.UpdateTask"

	// Prepare the dynamic update query based on the provided fields
	queryUpdate := "UPDATE tasks SET updated_at = $1"
	queryParams := []interface{}{time.Now().UTC()}

	// Add fields to the query
	if task.Title != "" {
		queryUpdate += ", title = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.Title)
	}
	if task.Description != "" {
		queryUpdate += ", description = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.Description)
	}
	if task.StartDate != nil {
		queryUpdate += ", start_date = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.StartDate)
	}
	if task.Deadline != nil {
		queryUpdate += ", deadline = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.Deadline)
	}

	// Add condition for the specific user ID
	queryUpdate += " WHERE id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, task.ID)

	queryUpdate += " AND user_id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, task.UserID)

	// Execute the update query
	result, err := s.Exec(ctx, queryUpdate, queryParams...)
	if err != nil {
		return fmt.Errorf("%s: failed to update task: %w", op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return c.ErrTaskNotFound
	}

	return nil
}

func (s *TaskStorage) UpdateTaskTime(ctx context.Context, task models.Task) error {
	const (
		op = "task.storage.UpdateTaskTime"

		querySelectStatus = `SELECT id FROM statuses WHERE status_name = $1`
	)

	// Get the status ID for the planned status
	var status string

	err := s.QueryRow(ctx, querySelectStatus, c.StatusPlanned).Scan(&status)
	if err != nil {
		return fmt.Errorf("%s: failed to get status: %w", op, err)
	}

	// Prepare the dynamic update query based on the provided fields
	queryUpdate := "UPDATE tasks SET updated_at = $1"
	queryParams := []interface{}{time.Now().UTC()}

	// Add status ID to the query
	queryUpdate += ", status_id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, status)

	// Add time fields to the query
	if task.StartTime != nil {
		queryUpdate += ", start_time = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.StartTime)
	}
	if task.EndTime != nil {
		queryUpdate += ", end_time = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.EndTime)
	}

	// Add condition for the specific user ID
	queryUpdate += " WHERE id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, task.ID)

	queryUpdate += " AND user_id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, task.UserID)

	// Execute the update query
	result, err := s.Exec(ctx, queryUpdate, queryParams...)
	if err != nil {
		return fmt.Errorf("%s: failed to update task: %w", op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return c.ErrTaskNotFound
	}

	return nil
}

func (s *TaskStorage) CompleteTask(ctx context.Context, taskID, userID string) error {
	const (
		op = "task.storage.CompleteTask"

		querySelectStatus = `SELECT id FROM statuses WHERE status_name = $1`

		queryUpdate = `
			UPDATE tasks
			SET
				status_id = $1,
				updated_at = $2
			WHERE id = $3
			AND user_id = $4
			AND deleted_at IS NULL
			RETURNING id`
	)

	var status string

	err := s.QueryRow(ctx, querySelectStatus, c.StatusCompleted).Scan(&status)
	if err != nil {
		return fmt.Errorf("%s: failed to get status: %w", op, err)
	}

	result, err := s.Exec(
		ctx,
		queryUpdate,
		status,
		time.Now(),
		taskID,
		userID)
	if err != nil {
		return fmt.Errorf("%s: failed to update task: %w", op, err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return c.ErrTaskNotFound
	}

	return nil
}

func (s *TaskStorage) DeleteTask(ctx context.Context, taskID, userID string) error {
	const (
		op = "task.storage.DeleteTask"

		querySelectStatus = `SELECT id FROM statuses WHERE status_name = $1`

		queryDeleteTask = `
			UPDATE tasks
			SET
			    status_id = $1,
				deleted_at = $2
			WHERE id = $3
			AND user_id = $4
			AND deleted_at IS NULL`
	)

	var status string

	err := s.QueryRow(ctx, querySelectStatus, c.StatusDeleted).Scan(&status)
	if err != nil {
		return fmt.Errorf("%s: failed to get status: %w", op, err)
	}

	_, err = s.Exec(
		ctx,
		queryDeleteTask,
		status,
		time.Now(),
		taskID,
		userID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return c.ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("%s: failed to delete task: %w", op, err)
	}

	return nil
}
