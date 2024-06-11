// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: task.sql

package sqlc

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const createTask = `-- name: CreateTask :exec
INSERT INTO tasks (
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
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
`

type CreateTaskParams struct {
	ID          string             `db:"id"`
	Title       string             `db:"title"`
	Description pgtype.Text        `db:"description"`
	StartDate   pgtype.Timestamptz `db:"start_date"`
	Deadline    pgtype.Timestamptz `db:"deadline"`
	StartTime   pgtype.Timestamptz `db:"start_time"`
	EndTime     pgtype.Timestamptz `db:"end_time"`
	StatusID    int32              `db:"status_id"`
	ListID      string             `db:"list_id"`
	HeadingID   string             `db:"heading_id"`
	UserID      string             `db:"user_id"`
	UpdatedAt   time.Time          `db:"updated_at"`
}

func (q *Queries) CreateTask(ctx context.Context, arg CreateTaskParams) error {
	_, err := q.db.Exec(ctx, createTask,
		arg.ID,
		arg.Title,
		arg.Description,
		arg.StartDate,
		arg.Deadline,
		arg.StartTime,
		arg.EndTime,
		arg.StatusID,
		arg.ListID,
		arg.HeadingID,
		arg.UserID,
		arg.UpdatedAt,
	)
	return err
}

const getArchivedTasks = `-- name: GetArchivedTasks :many
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
        ttv.tags as tags,
        t.updated_at,
        t.deleted_at
    FROM tasks t
        LEFT JOIN task_tags_view ttv
            ON t.id = ttv.task_id
    WHERE t.user_id = $1
      AND t.status_id = (
        SELECT id
        FROM statuses
        WHERE statuses.title = $3::varchar
        )
      AND (t.deleted_at IS NOT NULL
               OR (DATE_TRUNC('month', t.updated_at) > $4::timestamptz AND t.deleted_at IS NOT NULL)
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
ORDER BY month DESC
LIMIT $2
`

type GetArchivedTasksParams struct {
	UserID      string             `db:"user_id"`
	Limit       int32              `db:"limit"`
	StatusTitle string             `db:"status_title"`
	AfterMonth  pgtype.Timestamptz `db:"after_month"`
}

type GetArchivedTasksRow struct {
	Month pgtype.Interval `db:"month"`
	Tasks []byte          `db:"tasks"`
}

func (q *Queries) GetArchivedTasks(ctx context.Context, arg GetArchivedTasksParams) ([]GetArchivedTasksRow, error) {
	rows, err := q.db.Query(ctx, getArchivedTasks,
		arg.UserID,
		arg.Limit,
		arg.StatusTitle,
		arg.AfterMonth,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetArchivedTasksRow{}
	for rows.Next() {
		var i GetArchivedTasksRow
		if err := rows.Scan(&i.Month, &i.Tasks); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getCompletedTasks = `-- name: GetCompletedTasks :many
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
                            'updated_at', t.updated_at
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
        ttv.tags as tags,
        t.updated_at
    FROM tasks t
        LEFT JOIN task_tags_view ttv
            ON t.id = ttv.task_id
    WHERE t.user_id = $1
      AND t.status_id = (
      SELECT id
      FROM statuses
      WHERE statuses.title = $3::varchar
      )
      AND (t.deleted_at IS NULL
               OR (DATE_TRUNC('month', t.updated_at) > $4::timestamptz AND t.deleted_at IS NULL))
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
        t.updated_at
    ) t
GROUP BY month
ORDER BY month
LIMIT $2
`

type GetCompletedTasksParams struct {
	UserID      string             `db:"user_id"`
	Limit       int32              `db:"limit"`
	StatusTitle string             `db:"status_title"`
	AfterDate   pgtype.Timestamptz `db:"after_date"`
}

type GetCompletedTasksRow struct {
	Month pgtype.Interval `db:"month"`
	Tasks []byte          `db:"tasks"`
}

func (q *Queries) GetCompletedTasks(ctx context.Context, arg GetCompletedTasksParams) ([]GetCompletedTasksRow, error) {
	rows, err := q.db.Query(ctx, getCompletedTasks,
		arg.UserID,
		arg.Limit,
		arg.StatusTitle,
		arg.AfterDate,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetCompletedTasksRow{}
	for rows.Next() {
		var i GetCompletedTasksRow
		if err := rows.Scan(&i.Month, &i.Tasks); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getOverdueTasks = `-- name: GetOverdueTasks :many
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
                            'overdue', overdue,
                            'updated_at', t.updated_at
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
            ttv.tags as tags,
            CASE
                WHEN t.deadline <= CURRENT_DATE THEN TRUE
                ELSE FALSE END
                AS overdue,
            t.updated_at
        FROM tasks t
            LEFT JOIN task_tags_view ttv
                ON t.id = ttv.task_id
        WHERE t.user_id = $1
          AND t.deadline <= CURRENT_DATE
          AND (t.deleted_at IS NULL OR l.id > $3::varchar)
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
            t.updated_at
        ) t ON l.id = t.list_id
WHERE l.user_id = $1
GROUP BY l.id
ORDER BY l.id
LIMIT $2
`

type GetOverdueTasksParams struct {
	UserID  string `db:"user_id"`
	Limit   int32  `db:"limit"`
	AfterID string `db:"after_id"`
}

type GetOverdueTasksRow struct {
	ListID string `db:"list_id"`
	Tasks  []byte `db:"tasks"`
}

func (q *Queries) GetOverdueTasks(ctx context.Context, arg GetOverdueTasksParams) ([]GetOverdueTasksRow, error) {
	rows, err := q.db.Query(ctx, getOverdueTasks, arg.UserID, arg.Limit, arg.AfterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetOverdueTasksRow{}
	for rows.Next() {
		var i GetOverdueTasksRow
		if err := rows.Scan(&i.ListID, &i.Tasks); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTaskByID = `-- name: GetTaskByID :one
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
    t.updated_at,
    ttv.tags as tags,
    CASE
        WHEN t.deadline <= CURRENT_DATE THEN TRUE
        ELSE FALSE END
        AS overdue
FROM tasks t
    LEFT JOIN task_tags_view ttv
        ON t.id = ttv.task_id
WHERE t.id = $1
  AND t.user_id = $2
  AND t.deleted_at IS NULL
`

type GetTaskByIDParams struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`
}

type GetTaskByIDRow struct {
	ID          string             `db:"id"`
	Title       string             `db:"title"`
	Description pgtype.Text        `db:"description"`
	StartDate   pgtype.Timestamptz `db:"start_date"`
	Deadline    pgtype.Timestamptz `db:"deadline"`
	StartTime   pgtype.Timestamptz `db:"start_time"`
	EndTime     pgtype.Timestamptz `db:"end_time"`
	StatusID    int32              `db:"status_id"`
	ListID      string             `db:"list_id"`
	HeadingID   string             `db:"heading_id"`
	UpdatedAt   time.Time          `db:"updated_at"`
	Tags        interface{}        `db:"tags"`
	Overdue     bool               `db:"overdue"`
}

func (q *Queries) GetTaskByID(ctx context.Context, arg GetTaskByIDParams) (GetTaskByIDRow, error) {
	row := q.db.QueryRow(ctx, getTaskByID, arg.ID, arg.UserID)
	var i GetTaskByIDRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.Description,
		&i.StartDate,
		&i.Deadline,
		&i.StartTime,
		&i.EndTime,
		&i.StatusID,
		&i.ListID,
		&i.HeadingID,
		&i.UpdatedAt,
		&i.Tags,
		&i.Overdue,
	)
	return i, err
}

const getTaskStatusID = `-- name: GetTaskStatusID :one
SELECT id
FROM statuses
WHERE title = $1
`

func (q *Queries) GetTaskStatusID(ctx context.Context, title string) (int32, error) {
	row := q.db.QueryRow(ctx, getTaskStatusID, title)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const getTasksByListID = `-- name: GetTasksByListID :many
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
    t.updated_at,
    ttv.tags as tags,
    CASE
        WHEN t.deadline <= CURRENT_DATE THEN TRUE
        ELSE FALSE END
        AS overdue
FROM tasks t
    LEFT JOIN task_tags_view ttv
        ON t.id = ttv.task_id
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
ORDER BY t.id
`

type GetTasksByListIDParams struct {
	ListID string `db:"list_id"`
	UserID string `db:"user_id"`
}

type GetTasksByListIDRow struct {
	ID          string             `db:"id"`
	Title       string             `db:"title"`
	Description pgtype.Text        `db:"description"`
	StartDate   pgtype.Timestamptz `db:"start_date"`
	Deadline    pgtype.Timestamptz `db:"deadline"`
	StartTime   pgtype.Timestamptz `db:"start_time"`
	EndTime     pgtype.Timestamptz `db:"end_time"`
	StatusID    int32              `db:"status_id"`
	ListID      string             `db:"list_id"`
	HeadingID   string             `db:"heading_id"`
	UserID      string             `db:"user_id"`
	UpdatedAt   time.Time          `db:"updated_at"`
	Tags        interface{}        `db:"tags"`
	Overdue     bool               `db:"overdue"`
}

func (q *Queries) GetTasksByListID(ctx context.Context, arg GetTasksByListIDParams) ([]GetTasksByListIDRow, error) {
	rows, err := q.db.Query(ctx, getTasksByListID, arg.ListID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetTasksByListIDRow{}
	for rows.Next() {
		var i GetTasksByListIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.StartDate,
			&i.Deadline,
			&i.StartTime,
			&i.EndTime,
			&i.StatusID,
			&i.ListID,
			&i.HeadingID,
			&i.UserID,
			&i.UpdatedAt,
			&i.Tags,
			&i.Overdue,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTasksByUserID = `-- name: GetTasksByUserID :many
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
    t.updated_at,
    ttv.tags as tags,
    CASE
        WHEN t.deadline <= CURRENT_DATE THEN TRUE
        ELSE FALSE END
      AS overdue
FROM tasks t
    LEFT JOIN task_tags_view ttv
        ON t.id = ttv.task_id
WHERE t.user_id = $1
  AND t.deleted_at IS NULL
  AND (
      ($3::varchar IS NULL AND t.id > $3::varchar)
          OR ($3::varchar IS NOT NULL AND t.id > $3::varchar)
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
LIMIT $2
`

type GetTasksByUserIDParams struct {
	UserID  string `db:"user_id"`
	Limit   int32  `db:"limit"`
	AfterID string `db:"after_id"`
}

type GetTasksByUserIDRow struct {
	ID          string             `db:"id"`
	Title       string             `db:"title"`
	Description pgtype.Text        `db:"description"`
	StartDate   pgtype.Timestamptz `db:"start_date"`
	Deadline    pgtype.Timestamptz `db:"deadline"`
	StartTime   pgtype.Timestamptz `db:"start_time"`
	EndTime     pgtype.Timestamptz `db:"end_time"`
	StatusID    int32              `db:"status_id"`
	ListID      string             `db:"list_id"`
	HeadingID   string             `db:"heading_id"`
	UpdatedAt   time.Time          `db:"updated_at"`
	Tags        interface{}        `db:"tags"`
	Overdue     bool               `db:"overdue"`
}

func (q *Queries) GetTasksByUserID(ctx context.Context, arg GetTasksByUserIDParams) ([]GetTasksByUserIDRow, error) {
	rows, err := q.db.Query(ctx, getTasksByUserID, arg.UserID, arg.Limit, arg.AfterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetTasksByUserIDRow{}
	for rows.Next() {
		var i GetTasksByUserIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Description,
			&i.StartDate,
			&i.Deadline,
			&i.StartTime,
			&i.EndTime,
			&i.StatusID,
			&i.ListID,
			&i.HeadingID,
			&i.UpdatedAt,
			&i.Tags,
			&i.Overdue,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTasksForSomeday = `-- name: GetTasksForSomeday :many
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
                            'overdue', overdue,
                            'updated_at', t.updated_at
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
            ttv.tags as tags,
            CASE
                WHEN t.deadline <= CURRENT_DATE THEN TRUE
                ELSE FALSE END
                AS overdue,
            t.updated_at
        FROM tasks t
            LEFT JOIN task_tags_view ttv
                ON t.id = ttv.task_id
        WHERE t.user_id = $1
          AND t.start_date IS NULL
          AND t.deadline > CURRENT_DATE
          AND (t.deleted_at IS NULL OR l.id > $3::varchar)
        GROUP BY
            t.id,
            t.title,
            t.description,
            t.deadline,
            t.start_time,
            t.end_time,
            t.list_id,
            t.user_id,
            t.updated_at
        ) t ON l.id = t.list_id
WHERE l.user_id = $1
GROUP BY l.id
ORDER BY l.id
LIMIT $2
`

type GetTasksForSomedayParams struct {
	UserID  string `db:"user_id"`
	Limit   int32  `db:"limit"`
	AfterID string `db:"after_id"`
}

type GetTasksForSomedayRow struct {
	ListID string `db:"list_id"`
	Tasks  []byte `db:"tasks"`
}

func (q *Queries) GetTasksForSomeday(ctx context.Context, arg GetTasksForSomedayParams) ([]GetTasksForSomedayRow, error) {
	rows, err := q.db.Query(ctx, getTasksForSomeday, arg.UserID, arg.Limit, arg.AfterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetTasksForSomedayRow{}
	for rows.Next() {
		var i GetTasksForSomedayRow
		if err := rows.Scan(&i.ListID, &i.Tasks); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTasksForToday = `-- name: GetTasksForToday :many
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
                            'overdue', overdue,
                            'updated_at', t.updated_at
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
            ttv.tags as tags,
            CASE
                WHEN t.deadline <= CURRENT_DATE THEN TRUE
                ELSE FALSE END
                AS overdue,
            t.updated_at
        FROM tasks t
            LEFT JOIN task_tags_view ttv
                ON t.id = ttv.task_id
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
            t.updated_at
        ) t
        ON l.id = t.list_id
WHERE l.user_id = $1
GROUP BY l.id
ORDER BY l.id
`

type GetTasksForTodayRow struct {
	ListID string `db:"list_id"`
	Tasks  []byte `db:"tasks"`
}

func (q *Queries) GetTasksForToday(ctx context.Context, userID string) ([]GetTasksForTodayRow, error) {
	rows, err := q.db.Query(ctx, getTasksForToday, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetTasksForTodayRow{}
	for rows.Next() {
		var i GetTasksForTodayRow
		if err := rows.Scan(&i.ListID, &i.Tasks); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getTasksGroupedByHeadings = `-- name: GetTasksGroupedByHeadings :many
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
                            'overdue', overdue,
                            'updated_at', t.updated_at
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
            ttv.tags as tags,
            CASE
                WHEN t.deadline <= CURRENT_DATE THEN TRUE
                ELSE FALSE END
                AS overdue,
            t.updated_at
        FROM tasks t
            LEFT JOIN task_tags_view ttv
                ON t.id = ttv.task_id
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
            t.updated_at
    ) t
        ON h.id = t.heading_id
WHERE h.list_id = $1
  AND h.user_id = $2
GROUP BY h.id
ORDER BY h.id
`

type GetTasksGroupedByHeadingsParams struct {
	ListID string `db:"list_id"`
	UserID string `db:"user_id"`
}

type GetTasksGroupedByHeadingsRow struct {
	HeadingID string `db:"heading_id"`
	Tasks     []byte `db:"tasks"`
}

func (q *Queries) GetTasksGroupedByHeadings(ctx context.Context, arg GetTasksGroupedByHeadingsParams) ([]GetTasksGroupedByHeadingsRow, error) {
	rows, err := q.db.Query(ctx, getTasksGroupedByHeadings, arg.ListID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetTasksGroupedByHeadingsRow{}
	for rows.Next() {
		var i GetTasksGroupedByHeadingsRow
		if err := rows.Scan(&i.HeadingID, &i.Tasks); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUpcomingTasks = `-- name: GetUpcomingTasks :many
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
                            'updated_at', t.updated_at
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
        ttv.tags as tags,
        t.updated_at
    FROM tasks t
        LEFT JOIN task_tags_view ttv
            ON t.id = ttv.task_id
    WHERE t.user_id = $1
      AND (
          (t.start_date >= COALESCE($3::timestamptz, CURRENT_DATE + interval '1 day'))
              AND (t.deleted_at IS NULL)
              AND (COALESCE(t.start_date, $3::timestamptz) > $3::timestamptz)
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
        t.updated_at
    ) t
GROUP BY t.start_date
ORDER BY t.start_date
LIMIT $2
`

type GetUpcomingTasksParams struct {
	UserID    string             `db:"user_id"`
	Limit     int32              `db:"limit"`
	AfterDate pgtype.Timestamptz `db:"after_date"`
}

type GetUpcomingTasksRow struct {
	StartDate pgtype.Timestamptz `db:"start_date"`
	Tasks     []byte             `db:"tasks"`
}

func (q *Queries) GetUpcomingTasks(ctx context.Context, arg GetUpcomingTasksParams) ([]GetUpcomingTasksRow, error) {
	rows, err := q.db.Query(ctx, getUpcomingTasks, arg.UserID, arg.Limit, arg.AfterDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetUpcomingTasksRow{}
	for rows.Next() {
		var i GetUpcomingTasksRow
		if err := rows.Scan(&i.StartDate, &i.Tasks); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const markTaskAsArchived = `-- name: MarkTaskAsArchived :one
UPDATE tasks
SET status_id = $1, deleted_at = $2
WHERE id = $3
  AND user_id = $4
  AND deleted_at IS NULL
RETURNING id
`

type MarkTaskAsArchivedParams struct {
	StatusID  int32              `db:"status_id"`
	DeletedAt pgtype.Timestamptz `db:"deleted_at"`
	ID        string             `db:"id"`
	UserID    string             `db:"user_id"`
}

func (q *Queries) MarkTaskAsArchived(ctx context.Context, arg MarkTaskAsArchivedParams) (string, error) {
	row := q.db.QueryRow(ctx, markTaskAsArchived,
		arg.StatusID,
		arg.DeletedAt,
		arg.ID,
		arg.UserID,
	)
	var id string
	err := row.Scan(&id)
	return id, err
}

const markTaskAsCompleted = `-- name: MarkTaskAsCompleted :one
UPDATE tasks
SET	status_id = $1,
    updated_at = $2
WHERE id = $3
  AND user_id = $4
  AND deleted_at IS NULL
RETURNING id
`

type MarkTaskAsCompletedParams struct {
	StatusID  int32     `db:"status_id"`
	UpdatedAt time.Time `db:"updated_at"`
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
}

func (q *Queries) MarkTaskAsCompleted(ctx context.Context, arg MarkTaskAsCompletedParams) (string, error) {
	row := q.db.QueryRow(ctx, markTaskAsCompleted,
		arg.StatusID,
		arg.UpdatedAt,
		arg.ID,
		arg.UserID,
	)
	var id string
	err := row.Scan(&id)
	return id, err
}

const moveTaskToAnotherList = `-- name: MoveTaskToAnotherList :exec
UPDATE tasks
SET	list_id = $1,
    heading_id = $2,
    updated_at = $3
WHERE id = $4
  AND user_id = $5
  AND deleted_at IS NULL
`

type MoveTaskToAnotherListParams struct {
	ListID    string    `db:"list_id"`
	HeadingID string    `db:"heading_id"`
	UpdatedAt time.Time `db:"updated_at"`
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
}

func (q *Queries) MoveTaskToAnotherList(ctx context.Context, arg MoveTaskToAnotherListParams) error {
	_, err := q.db.Exec(ctx, moveTaskToAnotherList,
		arg.ListID,
		arg.HeadingID,
		arg.UpdatedAt,
		arg.ID,
		arg.UserID,
	)
	return err
}
