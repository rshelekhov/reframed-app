package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/pkg/constants/le"
	"strconv"
)

type TaskStorage struct {
	*pgxpool.Pool
}

func NewTaskStorage(pool *pgxpool.Pool) *TaskStorage {
	return &TaskStorage{Pool: pool}
}

func (s *TaskStorage) CreateTask(ctx context.Context, task model.Task) error {
	const (
		op = "task.storage.CreateTask"

		query = `
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
    			heading_id,
				user_id,
				updated_at
			)
			VALUES (
			    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	)

	_, err := s.Exec(
		ctx,
		query,
		task.ID,
		task.Title,
		task.Description,
		task.StartDate,
		task.Deadline,
		task.StartTime,
		task.EndTime,
		task.StatusID,
		task.ListID,
		task.HeadingID,
		task.UserID,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to insert new task: %w", op, err)
	}

	return nil
}

func (s *TaskStorage) GetTaskStatusID(ctx context.Context, status model.StatusName) (int, error) {
	const (
		op = "task.storage.GetTaskStatusID"

		query = `
			SELECT id
			FROM statuses
			WHERE status_name = $1`
	)

	var statusID int

	err := s.QueryRow(ctx, query, string(status)).Scan(&statusID)
	if err != nil {
		return 0, fmt.Errorf("%s: failed to get statusID: %w", op, err)
	}

	return statusID, nil
}

func (s *TaskStorage) GetTaskByID(ctx context.Context, taskID, userID string) (model.Task, error) {
	const (
		op = "task.storage.GetTaskByID"

		query = `
			SELECT
			    t.id,
				t.title,
				t.description,
				t.start_date,
				t.deadline,
				t.start_time,
				t.end_time,
				t.status_id,
				t.list_id,
				t.heading_id,
				array_agg(tg.title) AS tags,
				COALESCE(t.deadline <= CURRENT_DATE, false) AS overdue,
				t.updated_at
			FROM tasks t
				LEFT JOIN tasks_tags tt
				    ON t.id = tt.task_id
				LEFT JOIN tags tg
				    ON tt.tag_id = tg.id
			WHERE t.id = $1
			  AND t.user_id = $2
			  AND t.deleted_at IS NULL
			GROUP BY 
			    t.id,
			    t.title,
			    t.description,
			    t.start_date,
			    t.deadline,
			    t.start_time,
			    t.end_time,
			    t.status_id,
			    t.list_id,
			    t.heading_id,
			    t.updated_at`
	)

	var task model.Task

	err := s.QueryRow(
		ctx,
		query,
		taskID,
		userID,
	).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.StartDate,
		&task.Deadline,
		&task.StartTime,
		&task.EndTime,
		&task.StatusID,
		&task.ListID,
		&task.HeadingID,
		&task.Tags,
		&task.Overdue,
		&task.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Task{}, le.ErrTaskNotFound
	}
	if err != nil {
		return model.Task{}, fmt.Errorf("%s: failed to get task: %w", op, err)
	}

	return task, nil
}

func (s *TaskStorage) GetTasksByUserID(ctx context.Context, userID string, pgn model.Pagination) ([]model.Task, error) {
	const (
		op = "task.storage.GetTasksByUserID"

		query = `
			SELECT
				t.id,
				t.title,
				t.description,
				t.start_date,
				t.deadline,
				t.start_time,
				t.end_time,
				t.status_id,
				t.list_id,
				t.heading_id,
				ARRAY_AGG(tg.title) AS tags,
				COALESCE(t.deadline <= CURRENT_DATE, false) AS overdue,
				t.updated_at
			FROM tasks t
				LEFT JOIN tasks_tags tt
				    ON t.id = tt.task_id
				LEFT JOIN tags tg
				    ON tt.tag_id = tg.id
			WHERE t.user_id = $1
			  AND t.deleted_at IS NULL
			  AND (
			      ($2 IS NULL AND t.id > $2)
			      OR ($2 IS NOT NULL AND t.id > $2)
              )
			GROUP BY 
				t.id,
				t.title,
				t.description,
				t.start_date,
				t.deadline,
				t.start_time,
				t.end_time,
				t.status_id,
				t.list_id,
				t.heading_id,
				t.updated_at				
			ORDER BY t.id
			LIMIT $3`
	)

	var afterID interface{}
	if pgn.AfterID != "" {
		afterID = pgn.AfterID
	} else {
		afterID = nil
	}

	rows, err := s.Query(ctx, query, userID, afterID, pgn.Limit)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks: %w", op, err)
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		task := model.Task{}

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
			&task.HeadingID,
			&task.Tags,
			&task.Overdue,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(tasks) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return tasks, nil
}

func (s *TaskStorage) GetTasksByListID(ctx context.Context, listID, userID string) ([]model.Task, error) {
	const (
		op = "task.storage.GetTasksByListID"

		//query = `
		//	SELECT
		//		id,
		//		title,
		//		description,
		//		start_date,
		//		deadline,
		//		start_time,
		//		end_time,
		//		status_id,
		//		updated_at
		//	FROM tasks
		//	WHERE list_id = $1 AND user_id = $2 AND deleted_at IS NULL
		//	ORDER BY id DESC LIMIT $3 OFFSET $4`

		query = `
			SELECT
				t.id,
				t.title,
				t.description,
				t.start_date,
				t.deadline,
				t.start_time,
				t.end_time,
				t.status_id,
				t.list_id,
				t.heading_id,
				t.user_id,
				ARRAY_AGG(tg.title) AS tags,
				COALESCE(t.deadline <= CURRENT_DATE, false) AS overdue,
				t.updated_at
			FROM tasks t
				LEFT JOIN tasks_tags tt
				    ON t.id = tt.task_id
				LEFT JOIN tags tg
				    ON tt.tag_id = tg.id
			WHERE t.list_id = $1
			  AND t.user_id = $2
			  AND t.deleted_at IS NULL
			GROUP BY 
				t.id,
				t.title,
				t.description,
				t.start_date,
				t.deadline,
				t.start_time,
				t.end_time,
				t.status_id,
				t.heading_id,
				overdue,
				t.updated_at
			ORDER BY t.id`
	)

	rows, err := s.Query(ctx, query, listID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks: %w", op, err)
	}
	defer rows.Close()

	var tasks []model.Task

	for rows.Next() {
		task := model.Task{}

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
			&task.HeadingID,
			&task.UserID,
			&task.Tags,
			&task.Overdue,
			&task.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}

		tasks = append(tasks, task)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(tasks) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return tasks, nil
}

func (s *TaskStorage) GetTasksGroupedByHeadings(ctx context.Context, listID, userID string) ([]model.TaskGroup, error) {
	const (
		op = "task.storage.GetTasksGroupedByHeadings"

		query = `
			SELECT
				h.id AS heading_id,
				ARRAY_TO_JSON(
					ARRAY_AGG(
						JSON_BUILD_OBJECT(
							'id', t.id,
							'title', t.title,
							'description', t.description,
							'start_date', t.start_date,
							'deadline', t.deadline,
							'start_time', t.start_time,
							'end_time', t.end_time,
							'heading_id', t.heading_id,
							'user_id', t.user_id,
							'tags', tags,
							'overdue', t.deadline <= CURRENT_DATE,
							'updated_at', t.updated_at,
							'deleted_at', t.deleted_at
						)
					)
				) AS tasks
			FROM headings h
				LEFT JOIN (
					SELECT
						t.id,
						t.title,
						t.description,
						t.start_date,
						t.deadline,
						t.start_time,
						t.end_time,
						t.heading_id,
						t.user_id,
						ARRAY_AGG(tg.title) AS tags,
						t.updated_at,
						t.deleted_at
					FROM tasks t
						LEFT JOIN tasks_tags tt
						    ON t.id = tt.task_id
						LEFT JOIN tags tg
						    ON tt.tag_id = tg.id
					WHERE t.list_id = $1
					  AND t.user_id = $2
					  AND t.deleted_at IS NULL
					GROUP BY 
						t.id,
						t.title,
						t.description,
						t.start_date,
						t.deadline,
						t.start_time,
						t.end_time,
						t.heading_id,
						t.user_id,
						t.updated_at,
						t.deleted_at
				) t
				    ON h.id = t.heading_id
			WHERE h.list_id = $1
			  AND h.user_id = $2
			GROUP BY h.id
			ORDER BY h.id`
	)

	var taskGroups []model.TaskGroup

	rows, err := s.Query(ctx, query, listID, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskGroup model.TaskGroup
		var tasksJSON []byte

		err = rows.Scan(&taskGroup.HeadingID, &tasksJSON)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}

		var tasks []model.TaskResponseData

		err = json.Unmarshal(tasksJSON, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks: %w", op, err)
		}

		taskGroup.Tasks = tasks
		taskGroups = append(taskGroups, taskGroup)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(taskGroups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return taskGroups, nil
}

func (s *TaskStorage) GetTasksForToday(ctx context.Context, userID string) ([]model.TaskGroup, error) {
	const (
		op = "task.storage.GetTasksForToday"

		query = `
			SELECT
			    l.id AS list_id,
			    ARRAY_TO_JSON(
					ARRAY_AGG(
						JSON_BUILD_OBJECT(
							'id', t.id,
							'title', t.title,
							'description', t.description,
							'start_date', t.start_date,
							'deadline', t.deadline,
							'start_time', t.start_time,
							'end_time', t.end_time,
							'list_id', t.list_id,
							'user_id', t.user_id,
							'tags', tags,
							'overdue', t.deadline <= CURRENT_DATE,
							'updated_at', t.updated_at,
							'deleted_at', t.deleted_at
						)
					)
				) AS tasks
			FROM lists l
				LEFT JOIN (
					SELECT
						t.id,
						t.title,
						t.description,
						t.start_date,
						t.deadline,
						t.start_time,
						t.end_time,
						t.list_id,
						t.user_id,
						ARRAY_AGG(tg.title) AS tags,
						t.updated_at,
						t.deleted_at
					FROM tasks t
						LEFT JOIN tasks_tags tt
							ON t.id = tt.task_id
						LEFT JOIN tags tg
							ON tt.tag_id = tg.id
					WHERE t.user_id = $1
					  AND t.start_date = CURRENT_DATE
					  AND t.deleted_at IS NULL
					GROUP BY 
						t.id,
						t.title,
						t.description,
						t.start_date,
						t.deadline,
						t.start_time,
						t.end_time,
						t.list_id,
						t.user_id,
						t.updated_at,
						t.deleted_at
				) t
				    ON l.id = t.list_id
			WHERE l.user_id = $1
			GROUP BY l.id
			ORDER BY l.id`
	)

	var taskGroups []model.TaskGroup

	rows, err := s.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var taskGroup model.TaskGroup
		var tasksJSON []byte

		err = rows.Scan(&taskGroup.ListID, &tasksJSON)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}

		var tasks []model.TaskResponseData

		err = json.Unmarshal(tasksJSON, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks: %w", op, err)
		}

		taskGroup.Tasks = tasks
		taskGroups = append(taskGroups, taskGroup)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(taskGroups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return taskGroups, nil
}

func (s *TaskStorage) GetUpcomingTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	const (
		op = "task.storage.GetUpcomingTasks"

		query = `
			SELECT
			    t.start_date AS start_date,
			    ARRAY_TO_JSON(
					ARRAY_AGG(
						JSON_BUILD_OBJECT(
							'id', t.id,
							'title', t.title,
							'description', t.description,
							'start_date', t.start_date,
							'deadline', t.deadline,
							'start_time', t.start_time,
							'end_time', t.end_time,
							'list_id', t.list_id,
							'user_id', t.user_id,
							'tags', tags,
							'updated_at', t.updated_at,
							'deleted_at', t.deleted_at
						)
					)
				) AS tasks
			FROM (
				SELECT
					t.id,
					t.title,
					t.description,
					t.start_date,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					ARRAY_AGG(tg.title) AS tags,
					t.updated_at,
					t.deleted_at
				FROM tasks t
					LEFT JOIN tasks_tags tt
					    ON t.id = tt.task_id
					LEFT JOIN tags tg
					    ON tt.tag_id = tg.id
				WHERE t.user_id = $1
				  AND (
						(t.start_date >= COALESCE($2, CURRENT_DATE + interval '1 day'))
						AND (t.deleted_at IS NULL)
				  		AND (COALESCE(t.start_date, $2) > $2)
                  )
				GROUP BY 
					t.id,
					t.title,
					t.description,
					t.start_date,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					t.updated_at,
					t.deleted_at
			) t
			GROUP BY t.start_date
			ORDER BY t.start_date
			LIMIT $3`
	)

	var params []interface{}
	var afterDate interface{}

	if pgn.AfterID != "" {
		afterDate = pgn.AfterDate
	}

	params = append(params, userID, afterDate, pgn.Limit)

	rows, err := s.Query(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	defer rows.Close()

	var taskGroups []model.TaskGroup

	for rows.Next() {
		var taskGroup model.TaskGroup
		var tasksJSON []byte

		err = rows.Scan(&taskGroup.StartDate, &tasksJSON)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}

		var tasks []model.TaskResponseData

		err = json.Unmarshal(tasksJSON, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks: %w", op, err)
		}

		taskGroup.Tasks = tasks
		taskGroups = append(taskGroups, taskGroup)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(taskGroups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return taskGroups, nil
}

func (s *TaskStorage) GetOverdueTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	const (
		op = "task.storage.GetOverdueTasks"

		query = `
			SELECT
			    l.id AS list_id,
			    ARRAY_TO_JSON(
					ARRAY_AGG(
						JSON_BUILD_OBJECT(
							'id', t.id,
							'title', t.title,
							'description', t.description,
							'start_date', t.start_date,
							'deadline', t.deadline,
							'start_time', t.start_time,
							'end_time', t.end_time,
							'list_id', t.list_id,
							'user_id', t.user_id,
							'tags', tags,
							'overdue', t.deadline <= CURRENT_DATE,
							'updated_at', t.updated_at,
							'deleted_at', t.deleted_at
						)
					)
				) AS tasks
			FROM lists l
			LEFT JOIN (
				SELECT
					t.id,
					t.title,
					t.description,
					t.start_date,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					ARRAY_AGG(tg.title) AS tags,
					t.updated_at,
					t.deleted_at
				FROM tasks t
					LEFT JOIN tasks_tags tt
					    ON t.id = tt.task_id
					LEFT JOIN tags tg
					    ON tt.tag_id = tg.id
				WHERE t.user_id = $1
				  AND t.deadline <= CURRENT_DATE
				  AND (t.deleted_at IS NULL OR l.id > $2)
				GROUP BY 
					t.id,
					t.title,
					t.description,
					t.start_date,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					t.updated_at,
					t.deleted_at
			) t ON l.id = t.list_id
			WHERE l.user_id = $1
			GROUP BY l.id
			ORDER BY l.id
			LIMIT $3`
	)

	rows, err := s.Query(ctx, query, userID, pgn.AfterID, pgn.Limit)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	defer rows.Close()

	var taskGroups []model.TaskGroup

	for rows.Next() {
		var taskGroup model.TaskGroup
		var tasksJSON []byte

		err = rows.Scan(&taskGroup.ListID, &tasksJSON)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}

		var tasks []model.TaskResponseData

		err = json.Unmarshal(tasksJSON, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks: %w", op, err)
		}

		taskGroup.Tasks = tasks
		taskGroups = append(taskGroups, taskGroup)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(taskGroups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return taskGroups, nil
}

func (s *TaskStorage) GetTasksForSomeday(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	const (
		op = "task.storage.GetTasksForSomeday"

		query = `
			SELECT
			    l.id AS list_id,
			    ARRAY_TO_JSON(
					ARRAY_AGG(
						JSON_BUILD_OBJECT(
							'id', t.id,
							'title', t.title,
							'description', t.description,
							'deadline', t.deadline,
							'start_time', t.start_time,
							'end_time', t.end_time,
							'list_id', t.list_id,
							'user_id', t.user_id,
							'tags', tags,
							'overdue', t.deadline <= CURRENT_DATE,
							'updated_at', t.updated_at,
							'deleted_at', t.deleted_at
						)
					)
				) AS tasks
			FROM lists l
			LEFT JOIN (
				SELECT
					t.id,
					t.title,
					t.description,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					ARRAY_AGG(tg.title) AS tags,
					t.updated_at,
					t.deleted_at
				FROM tasks t
					LEFT JOIN tasks_tags tt
					    ON t.id = tt.task_id
					LEFT JOIN tags tg
					    ON tt.tag_id = tg.id
				WHERE t.user_id = $1
				  AND t.start_date IS NULL
				  AND t.deadline > CURRENT_DATE
				  AND (t.deleted_at IS NULL OR l.id > $2)
				GROUP BY 
					t.id,
					t.title,
					t.description,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					t.updated_at,
					t.deleted_at
			) t ON l.id = t.list_id
			WHERE l.user_id = $1
			GROUP BY l.id
			ORDER BY l.id
			LIMIT $3`
	)

	rows, err := s.Query(ctx, query, userID, pgn.AfterID, pgn.Limit)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	defer rows.Close()

	var taskGroups []model.TaskGroup

	for rows.Next() {
		var taskGroup model.TaskGroup
		var tasksJSON []byte

		err = rows.Scan(&taskGroup.ListID, &tasksJSON)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}

		var tasks []model.TaskResponseData

		err = json.Unmarshal(tasksJSON, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks: %w", op, err)
		}

		taskGroup.Tasks = tasks
		taskGroups = append(taskGroups, taskGroup)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(taskGroups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return taskGroups, nil
}

func (s *TaskStorage) GetCompletedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	const (
		op    = "task.storage.GetCompletedTasks"
		query = `
			SELECT
			    DATE_TRUNC('month', t.updated_at) AS month,
			    ARRAY_TO_JSON(
					ARRAY_AGG(
						JSON_BUILD_OBJECT(
							'id', t.id,
							'title', t.title,
							'description', t.description,
							'start_date', t.start_date,
							'deadline', t.deadline,
							'start_time', t.start_time,
							'end_time', t.end_time,
							'list_id', t.list_id,
							'user_id', t.user_id,
							'tags', tags,
							'updated_at', t.updated_at,
							'deleted_at', t.deleted_at
						)
					)
				) AS tasks
			FROM (
				SELECT
					t.id,
					t.title,
					t.description,
					t.start_date,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					ARRAY_AGG(tg.title) AS tags,
					t.updated_at,
					t.deleted_at
				FROM tasks t
					LEFT JOIN tasks_tags tt
					    ON t.id = tt.task_id
					LEFT JOIN tags tg
					    ON tt.tag_id = tg.id
				WHERE t.user_id = $1
				  AND t.status_id = (
				  		SELECT id
				    	FROM statuses
				    	WHERE status_name = $2
				  ) 
				  AND (t.deleted_at IS NULL
				           OR (DATE_TRUNC('month', t.updated_at) > $3 AND t.deleted_at IS NULL))
				GROUP BY 
					t.id,
					t.title,
					t.description,
					t.start_date,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					t.updated_at,
					t.deleted_at
			) t
			GROUP BY month
			ORDER BY month
			LIMIT $4`
	)

	queryParams := []interface{}{userID, model.StatusCompleted}

	if pgn.AfterID != "" {
		queryParams = append(queryParams, pgn.AfterDate)
	} else {
		queryParams = append(queryParams, nil)
	}

	queryParams = append(queryParams, pgn.Limit)

	rows, err := s.Query(ctx, query, queryParams...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	defer rows.Close()

	var taskGroups []model.TaskGroup

	for rows.Next() {
		var taskGroup model.TaskGroup
		var tasksJSON []byte

		err = rows.Scan(&taskGroup.Month, &tasksJSON)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}

		var tasks []model.TaskResponseData

		err = json.Unmarshal(tasksJSON, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks: %w", op, err)
		}

		taskGroup.Tasks = tasks
		taskGroups = append(taskGroups, taskGroup)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(taskGroups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return taskGroups, nil
}

func (s *TaskStorage) GetArchivedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	const (
		op = "task.storage.GetArchivedTasks"

		query = `
			SELECT
			    DATE_TRUNC('month', t.updated_at) AS month,
			    ARRAY_TO_JSON(
					ARRAY_AGG(
						JSON_BUILD_OBJECT(
							'id', t.id,
							'title', t.title,
							'description', t.description,
							'start_date', t.start_date,
							'deadline', t.deadline,
							'start_time', t.start_time,
							'end_time', t.end_time,
							'list_id', t.list_id,
							'user_id', t.user_id,
							'tags', tags,
							'updated_at', t.updated_at,
							'deleted_at', t.deleted_at
						)
					)
				) AS tasks
			FROM (
				SELECT
					t.id,
					t.title,
					t.description,
					t.start_date,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					ARRAY_AGG(tg.title) AS tags,
					t.updated_at,
					t.deleted_at
				FROM tasks t
				LEFT JOIN tasks_tags tt ON t.id = tt.task_id
				LEFT JOIN tags tg ON tt.tag_id = tg.id
				WHERE t.user_id = $1
				  AND t.status_id = (
						SELECT id
				    	FROM statuses
				    	WHERE status_name = $2
				  )
				  AND (t.deleted_at IS NULL 
						OR (DATE_TRUNC('month', t.updated_at) > $3 AND t.deleted_at IS NULL)
				  )
				GROUP BY 
					t.id,
					t.title,
					t.description,
					t.start_date,
					t.deadline,
					t.start_time,
					t.end_time,
					t.list_id,
					t.user_id,
					t.updated_at,
					t.deleted_at
			) t
			GROUP BY month
			ORDER BY month
			LIMIT $4`
	)

	queryParams := []interface{}{userID, model.StatusCompleted}

	if pgn.AfterID != "" {
		queryParams = append(queryParams, pgn.AfterDate)
	} else {
		queryParams = append(queryParams, nil)
	}

	queryParams = append(queryParams, pgn.Limit)

	rows, err := s.Query(ctx, query, queryParams...)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get tasks groups: %w", op, err)
	}
	defer rows.Close()

	var taskGroups []model.TaskGroup

	for rows.Next() {
		var taskGroup model.TaskGroup
		var tasksJSON []byte

		err = rows.Scan(&taskGroup.Month, &tasksJSON)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to scan task: %w", op, err)
		}

		var tasks []model.TaskResponseData

		err = json.Unmarshal(tasksJSON, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks: %w", op, err)
		}

		taskGroup.Tasks = tasks
		taskGroups = append(taskGroups, taskGroup)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: failed to iterate rows: %w", op, err)
	}

	if len(taskGroups) == 0 {
		return nil, le.ErrNoTasksFound
	}

	return taskGroups, nil
}

func (s *TaskStorage) UpdateTask(ctx context.Context, task model.Task) error {
	const (
		op = "task.storage.UpdateTask"

		queryGetHeadingID = `
			SELECT heading_id
			FROM tasks
			WHERE id = $1
			  AND user_id = $2`
	)

	var headingID string

	err := s.QueryRow(ctx, queryGetHeadingID, task.ID, task.UserID).Scan(&headingID)
	if err != nil {
		return fmt.Errorf("%s: failed to get heading ID: %w", op, err)
	}

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
	if task.HeadingID != headingID {
		queryUpdate += ", heading_id = $" + strconv.Itoa(len(queryParams)+1)
		queryParams = append(queryParams, task.HeadingID)
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

	// Get the statusID ID for the planned status
	var statusID string

	// Prepare the dynamic update query based on the provided fields
	queryUpdate := "UPDATE tasks SET updated_at = $1"
	queryParams := []interface{}{task.UpdatedAt}

	// Add time fields to the query
	if !task.StartTime.IsZero() && !task.EndTime.IsZero() {
		queryUpdate += ", start_time = $" + strconv.Itoa(len(queryParams)+1) + ", end_time = $" + strconv.Itoa(len(queryParams)+2)
		queryParams = append(queryParams, task.StartTime, task.EndTime)
	} else if task.StartTime.IsZero() && task.EndTime.IsZero() {
		queryUpdate += ", start_time = NULL, end_time = NULL"
	} else {
		return le.ErrInvalidTaskTimeRange
	}

	// Add statusID ID to the query
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
		return fmt.Errorf("%s: failed to update task: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return le.ErrTaskNotFound
	}

	return nil
}

func (s *TaskStorage) MoveTaskToAnotherList(ctx context.Context, task model.Task) error {
	const (
		op = "task.storage.MoveTaskToAnotherList"

		query = `
			UPDATE tasks
			SET	list_id = $1,
				heading_id = $2,
				updated_at = $3
			WHERE id = $4
			  AND user_id = $5
			  AND deleted_at IS NULL`
	)

	result, err := s.Exec(
		ctx,
		query,
		task.ListID,
		task.HeadingID,
		task.UpdatedAt,
		task.ID,
		task.UserID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to move task: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return le.ErrTaskNotFound
	}

	return nil
}

func (s *TaskStorage) CompleteTask(ctx context.Context, task model.Task) error {
	const (
		op = "task.storage.CompleteTask"

		query = `
			UPDATE tasks
			SET	status_id = $1, updated_at = $2
			WHERE id = $3
			  AND user_id = $4
			  AND deleted_at IS NULL`
	)

	result, err := s.Exec(
		ctx,
		query,
		task.StatusID,
		task.UpdatedAt,
		task.ID,
		task.UserID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to update task: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return le.ErrTaskNotFound
	}

	return nil
}

func (s *TaskStorage) ArchiveTask(ctx context.Context, task model.Task) error {
	const (
		op = "task.storage.ArchiveTask"

		query = `
			UPDATE tasks
			SET status_id = $1, deleted_at = $2
			WHERE id = $3
			  AND user_id = $4
			  AND deleted_at IS NULL`
	)

	result, err := s.Exec(
		ctx,
		query,
		task.StatusID,
		task.UpdatedAt,
		task.ID,
		task.UserID,
	)
	if err != nil {
		return fmt.Errorf("%s: failed to delete task: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return le.ErrTaskNotFound
	}

	return nil
}
