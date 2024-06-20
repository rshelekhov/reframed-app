package postgres

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
	"github.com/rshelekhov/reframed/internal/storage/postgres/sqlc"
)

type TaskStorage struct {
	*pgxpool.Pool
	*sqlc.Queries
}

func NewTaskStorage(pool *pgxpool.Pool) port.TaskStorage {
	return &TaskStorage{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
}

func (s *TaskStorage) Transaction(ctx context.Context, fn func(storage port.TaskStorage) error) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				err = fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
			}
		} else {
			err = tx.Commit(ctx)
		}
	}()

	return fn(s)
}

func (s *TaskStorage) CreateTask(ctx context.Context, task model.Task) error {
	const op = "task.storage.CreateTask"

	taskParams := sqlc.CreateTaskParams{
		ID:        task.ID,
		Title:     task.Title,
		StatusID:  int32(task.StatusID),
		ListID:    task.ListID,
		HeadingID: task.HeadingID,
		UserID:    task.UserID,
		UpdatedAt: task.UpdatedAt,
	}
	if task.Description != "" {
		taskParams.Description = pgtype.Text{
			String: task.Description,
			Valid:  true,
		}
	}
	if !task.StartDate.IsZero() {
		taskParams.StartDate = pgtype.Timestamptz{
			Time:  task.StartDate,
			Valid: true,
		}
	}
	if !task.Deadline.IsZero() {
		taskParams.Deadline = pgtype.Timestamptz{
			Time:  task.Deadline,
			Valid: true,
		}
	}

	if !task.StartTime.IsZero() {
		taskParams.StartTime = pgtype.Timestamptz{
			Time:  task.StartTime,
			Valid: true,
		}
	}
	if !task.EndTime.IsZero() {
		taskParams.EndTime = pgtype.Timestamptz{
			Time:  task.EndTime,
			Valid: true,
		}
	}

	if err := s.Queries.CreateTask(ctx, taskParams); err != nil {
		return fmt.Errorf("%s: failed to insert new task: %w", op, err)
	}
	return nil
}

func (s *TaskStorage) GetTaskStatusID(ctx context.Context, status model.StatusName) (int, error) {
	const op = "task.storage.GetTaskStatusID"

	statusID, err := s.Queries.GetTaskStatusID(ctx, status.String())

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return 0, le.ErrTaskStatusIDNotFound
	case err != nil:
		return 0, fmt.Errorf("%s: failed to get statusID: %w", op, err)
	default:
		return int(statusID), nil
	}
}

func (s *TaskStorage) GetTaskByID(ctx context.Context, taskID, userID string) (model.Task, error) {
	const op = "task.storage.GetTaskByID"

	task, err := s.Queries.GetTaskByID(ctx, sqlc.GetTaskByIDParams{
		ID:     taskID,
		UserID: userID,
	})

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return model.Task{}, le.ErrTaskNotFound
	case err != nil:
		return model.Task{}, fmt.Errorf("%s: failed to get task: %w", op, err)
	}

	taskResp := model.Task{
		ID:        task.ID,
		Title:     task.Title,
		StatusID:  int(task.StatusID),
		ListID:    task.ListID,
		HeadingID: task.HeadingID,
		UpdatedAt: task.UpdatedAt,
		Overdue:   task.Overdue,
	}
	if task.Description.Valid {
		taskResp.Description = task.Description.String
	}
	if task.StartDate.Valid {
		taskResp.StartDate = task.StartDate.Time
	}
	if task.Deadline.Valid {
		taskResp.Deadline = task.Deadline.Time
	}
	if task.StartTime.Valid {
		taskResp.StartTime = task.StartTime.Time
	}
	if task.EndTime.Valid {
		taskResp.EndTime = task.EndTime.Time
	}

	if task.Tags != nil {
		tagsArray, ok := task.Tags.([]interface{})
		if ok {
			tags := make([]string, 0, len(tagsArray))

			for _, tag := range tagsArray {
				if t, ok := tag.(string); ok {
					tags = append(tags, t)
				}
			}

			taskResp.Tags = tags
		}
	}

	return taskResp, nil
}

func (s *TaskStorage) GetTasksByUserID(ctx context.Context, userID string, pgn model.Pagination) ([]model.Task, error) {
	const op = "task.storage.GetTasksByUserID"

	tasksRaw, err := s.Queries.GetTasksByUserID(ctx, sqlc.GetTasksByUserIDParams{
		UserID:  userID,
		AfterID: pgn.AfterID,
		Limit:   pgn.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks: %w", op, err)
	}

	var tasks []interface{}
	for _, task := range tasksRaw {
		tasks = append(tasks, task)
	}

	tasksResp, err := transformTasks(tasks)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return tasksResp, nil
}

func (s *TaskStorage) GetTasksByListID(ctx context.Context, listID, userID string) ([]model.Task, error) {
	const op = "task.storage.GetTasksByListID"

	tasksRaw, err := s.Queries.GetTasksByListID(ctx, sqlc.GetTasksByListIDParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks: %w", op, err)
	}

	var tasks []interface{}
	for _, task := range tasksRaw {
		tasks = append(tasks, task)
	}

	tasksResp, err := transformTasks(tasks)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return tasksResp, nil
}

func transformTasks(tasks []interface{}) ([]model.Task, error) {
	var tasksResp []model.Task

	for _, task := range tasks {
		t, err := transformTask(task)
		if err != nil {
			return nil, err
		}

		tasksResp = append(tasksResp, t)
	}

	return tasksResp, nil
}

func transformTask(task interface{}) (model.Task, error) {
	switch t := task.(type) {
	case sqlc.GetTasksByUserIDRow:
		return transformGetTasksByUserIDRow(t)
	case sqlc.GetTasksByListIDRow:
		return transformGetTasksByListIDRow(t)
	default:
		return model.Task{}, errors.New("unsupported task type")
	}
}

func transformGetTasksByUserIDRow(task sqlc.GetTasksByUserIDRow) (model.Task, error) {
	t := model.Task{
		ID:        task.ID,
		Title:     task.Title,
		StatusID:  int(task.StatusID),
		ListID:    task.ListID,
		HeadingID: task.HeadingID,
		UpdatedAt: task.UpdatedAt,
		Overdue:   task.Overdue,
	}

	if task.Description.Valid {
		t.Description = task.Description.String
	}
	if task.StartDate.Valid {
		t.StartDate = task.StartDate.Time
	}
	if task.Deadline.Valid {
		t.Deadline = task.Deadline.Time
	}
	if task.StartTime.Valid {
		t.StartTime = task.StartTime.Time
	}
	if task.EndTime.Valid {
		t.EndTime = task.EndTime.Time
	}

	if task.Tags != nil {
		tags, err := transformTags(task.Tags)
		if err != nil {
			return model.Task{}, err
		}

		t.Tags = tags
	}

	return t, nil
}

func transformGetTasksByListIDRow(task sqlc.GetTasksByListIDRow) (model.Task, error) {
	t := model.Task{
		ID:        task.ID,
		Title:     task.Title,
		StatusID:  int(task.StatusID),
		ListID:    task.ListID,
		HeadingID: task.HeadingID,
		UpdatedAt: task.UpdatedAt,
		Overdue:   task.Overdue,
	}

	if task.Description.Valid {
		t.Description = task.Description.String
	}
	if task.StartDate.Valid {
		t.StartDate = task.StartDate.Time
	}
	if task.Deadline.Valid {
		t.Deadline = task.Deadline.Time
	}
	if task.StartTime.Valid {
		t.StartTime = task.StartTime.Time
	}
	if task.EndTime.Valid {
		t.EndTime = task.EndTime.Time
	}

	if task.Tags != nil {
		tags, err := transformTags(task.Tags)
		if err != nil {
			return model.Task{}, err
		}

		t.Tags = tags
	}

	return t, nil
}

func transformTags(tags interface{}) ([]string, error) {
	tagsArray, ok := tags.([]interface{})
	if !ok {
		return nil, errors.New("invalid tags format")
	}

	var transformedTags []string

	for _, tag := range tagsArray {
		if t, ok := tag.(string); ok {
			transformedTags = append(transformedTags, t)
		}
	}
	return transformedTags, nil
}

func (s *TaskStorage) GetTasksGroupedByHeadings(ctx context.Context, listID, userID string) ([]model.TaskGroupRaw, error) {
	const op = "task.storage.GetTasksGroupedByHeadings"

	groups, err := s.Queries.GetTasksGroupedByHeadings(ctx, sqlc.GetTasksGroupedByHeadingsParams{
		ListID: listID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	if len(groups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	var groupsRaw []model.TaskGroupRaw

	for _, group := range groups {
		var taskGroup model.TaskGroupRaw

		taskGroup.HeadingID = group.HeadingID
		taskGroup.Tasks = group.Tasks

		groupsRaw = append(groupsRaw, taskGroup)
	}

	return groupsRaw, nil
}

func (s *TaskStorage) GetTasksForToday(ctx context.Context, userID string) ([]model.TaskGroupRaw, error) {
	const op = "task.storage.GetTasksForToday"

	groups, err := s.Queries.GetTasksForToday(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	if len(groups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	var groupsRaw []model.TaskGroupRaw

	for _, group := range groups {
		var taskGroup model.TaskGroupRaw

		taskGroup.ListID = group.ListID
		taskGroup.Tasks = group.Tasks

		groupsRaw = append(groupsRaw, taskGroup)
	}

	return groupsRaw, nil
}

func (s *TaskStorage) GetUpcomingTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroupRaw, error) {
	const op = "task.storage.GetUpcomingTasks"

	groups, err := s.Queries.GetUpcomingTasks(ctx, sqlc.GetUpcomingTasksParams{
		UserID: userID,
		AfterDate: pgtype.Timestamptz{
			Valid: true,
			Time:  pgn.AfterDate,
		},
		Limit: pgn.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	if len(groups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	var groupsRaw []model.TaskGroupRaw

	for _, group := range groups {
		var taskGroup model.TaskGroupRaw

		taskGroup.StartDate = group.StartDate.Time
		taskGroup.Tasks = group.Tasks

		groupsRaw = append(groupsRaw, taskGroup)
	}

	return groupsRaw, nil
}

func (s *TaskStorage) GetOverdueTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroupRaw, error) {
	const op = "task.storage.GetOverdueTasks"

	groups, err := s.Queries.GetOverdueTasks(ctx, sqlc.GetOverdueTasksParams{
		UserID:  userID,
		Limit:   pgn.Limit,
		AfterID: pgn.AfterID,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	if len(groups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	var groupsRaw []model.TaskGroupRaw

	for _, group := range groups {
		var taskGroup model.TaskGroupRaw

		taskGroup.ListID = group.ListID
		taskGroup.Tasks = group.Tasks

		groupsRaw = append(groupsRaw, taskGroup)
	}

	return groupsRaw, nil
}

func (s *TaskStorage) GetTasksForSomeday(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroupRaw, error) {
	const op = "task.storage.GetTasksForSomeday"

	groups, err := s.Queries.GetUpcomingTasks(ctx, sqlc.GetUpcomingTasksParams{
		UserID: userID,
		AfterDate: pgtype.Timestamptz{
			Valid: true,
			Time:  pgn.AfterDate,
		},
		Limit: pgn.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	if len(groups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	var groupsRaw []model.TaskGroupRaw

	for _, group := range groups {
		var taskGroup model.TaskGroupRaw

		if group.StartDate.Valid {
			taskGroup.StartDate = group.StartDate.Time
		}

		taskGroup.Tasks = group.Tasks

		groupsRaw = append(groupsRaw, taskGroup)
	}

	return groupsRaw, nil
}

func (s *TaskStorage) GetCompletedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroupRaw, error) {
	const op = "task.storage.GetCompletedTasks"

	groups, err := s.Queries.GetCompletedTasks(ctx, sqlc.GetCompletedTasksParams{
		UserID:      userID,
		Limit:       pgn.Limit,
		StatusTitle: model.StatusCompleted.String(),
		AfterDate: pgtype.Timestamptz{
			Valid: true,
			Time:  pgn.AfterDate,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	if len(groups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	var taskGroups []model.TaskGroupRaw

	for _, group := range groups {
		var taskGroup model.TaskGroupRaw

		//if group.Month.Valid {
		//	taskGroup.Month = group.Month.Time
		//}

		taskGroup.Tasks = group.Tasks

		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, nil
}

func (s *TaskStorage) GetArchivedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroupRaw, error) {
	const op = "task.storage.GetArchivedTasks"

	var afterMonth time.Time

	if pgn.AfterDate.IsZero() {
		afterMonth = time.Now().Truncate(24 * time.Hour)
	} else {
		afterMonth = pgn.AfterDate
	}

	groups, err := s.Queries.GetArchivedTasks(ctx, sqlc.GetArchivedTasksParams{
		UserID:      userID,
		Limit:       pgn.Limit,
		StatusTitle: model.StatusArchived.String(),
		AfterMonth: pgtype.Timestamptz{
			Valid: true,
			Time:  afterMonth,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	if len(groups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	var taskGroups []model.TaskGroupRaw

	for _, group := range groups {
		var taskGroup model.TaskGroupRaw

		if group.Month.Valid {
			taskGroup.Month = group.Month.Time
		}

		taskGroup.Tasks = group.Tasks

		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, nil
}

func (s *TaskStorage) UpdateTask(ctx context.Context, task model.Task) error {
	const op = "task.storage.UpdateTask"

	// Prepare the dynamic update query based on the provided fields
	queryUpdate := "UPDATE tasks SET updated_at = $1"
	queryParams := []interface{}{task.UpdatedAt}

	// Add fields to the query
	if task.Title != "" {
		queryUpdate += ", title = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.Title)
	}
	if task.Description != "" {
		queryUpdate += ", description = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.Description)
	}
	if !task.StartDate.IsZero() {
		queryUpdate += ", start_date = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.StartDate)
	}
	if !task.Deadline.IsZero() {
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

	if result.RowsAffected() == 0 {
		return le.ErrTaskNotFound
	}

	return nil
}

func (s *TaskStorage) UpdateTaskTime(ctx context.Context, task model.Task) error {
	const op = "task.storage.UpdateTaskTime"

	// Prepare the dynamic update query based on the provided fields
	queryUpdate := "UPDATE tasks SET updated_at = $1"
	queryParams := []interface{}{task.UpdatedAt}

	// Add time fields to the query
	switch {
	case !task.StartTime.IsZero() && !task.EndTime.IsZero():
		queryUpdate += ", start_time = $" + strconv.Itoa(len(queryParams)+1) + ", end_time = $" + strconv.Itoa(len(queryParams)+2)
		queryParams = append(queryParams, task.StartTime, task.EndTime)
	case task.StartTime.IsZero() && task.EndTime.IsZero():
		queryUpdate += ", start_time = NULL, end_time = NULL"
	default:
		return le.ErrInvalidTaskTimeRange
	}

	// Add statusID to the query
	statusID := strconv.Itoa(task.StatusID)

	queryUpdate += ", status_id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, statusID)

	// Add condition for the specific user ID
	queryUpdate += " WHERE id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, task.ID)

	queryUpdate += " AND user_id = $" + strconv.Itoa(len(queryParams)+1)
	queryParams = append(queryParams, task.UserID)

	// Execute the update query
	result, err := s.Exec(ctx, queryUpdate, queryParams...)
	if err != nil {
		return fmt.Errorf("%s: failed to update task time: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return le.ErrTaskNotFound
	}

	return nil
}

func (s *TaskStorage) MoveTaskToAnotherList(ctx context.Context, task model.Task) error {
	const op = "task.storage.MoveTaskToAnotherList"

	_, err := s.Queries.MoveTaskToAnotherList(ctx, sqlc.MoveTaskToAnotherListParams{
		ListID:    task.ListID,
		HeadingID: task.HeadingID,
		UpdatedAt: task.UpdatedAt,
		ID:        task.ID,
		UserID:    task.UserID,
	})

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return le.ErrTaskNotFound
	case err != nil:
		return fmt.Errorf("%s: failed to move task to another list: %w", op, err)
	default:
		return nil
	}
}

func (s *TaskStorage) MoveTaskToAnotherHeading(ctx context.Context, task model.Task) error {
	const op = "task.storage.MoveTaskToAnotherHeading"

	_, err := s.Queries.MoveTaskToAnotherHeading(ctx, sqlc.MoveTaskToAnotherHeadingParams{
		HeadingID: task.HeadingID,
		UpdatedAt: task.UpdatedAt,
		ID:        task.ID,
		UserID:    task.UserID,
	})

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return le.ErrTaskNotFound
	case err != nil:
		return fmt.Errorf("%s: failed to move task to another heading: %w", op, err)
	default:
		return nil
	}
}

func (s *TaskStorage) MarkAsCompleted(ctx context.Context, task model.Task) error {
	const op = "task.storage.MarkAsCompleted"

	_, err := s.Queries.MarkTaskAsCompleted(ctx, sqlc.MarkTaskAsCompletedParams{
		StatusID:  int32(task.StatusID),
		UpdatedAt: task.UpdatedAt,
		ID:        task.ID,
		UserID:    task.UserID,
	})

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return le.ErrTaskNotFound
	case err != nil:
		return fmt.Errorf("%s: failed to mark task as completed: %w", op, err)
	default:
		return nil
	}
}

func (s *TaskStorage) MarkAsArchived(ctx context.Context, task model.Task) error {
	const op = "task.storage.MarkAsArchived"

	_, err := s.Queries.MarkTaskAsArchived(ctx, sqlc.MarkTaskAsArchivedParams{
		StatusID: int32(task.StatusID),
		DeletedAt: pgtype.Timestamptz{
			Valid: true,
			Time:  task.DeletedAt,
		},
		ID:     task.ID,
		UserID: task.UserID,
	})

	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return le.ErrTaskNotFound
	case err != nil:
		return fmt.Errorf("%s: failed to mark task as archived: %w", op, err)
	default:
		return nil
	}
}
