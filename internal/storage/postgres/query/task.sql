-- name: CreateTask :exec
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
);

-- name: GetTaskStatusID :one
SELECT id
FROM statuses
WHERE title = $1;

-- name: GetTaskByID :one
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
  AND t.deleted_at IS NULL;

-- name: GetTasksByUserID :many
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
      (@after_id::varchar IS NULL AND t.id > @after_id::varchar)
          OR (@after_id::varchar IS NOT NULL AND t.id > @after_id::varchar)
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
    t.updated_at,
    ttv.tags
ORDER BY t.id
LIMIT $2;

-- name: GetTasksByListID :many
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
ORDER BY t.id;

-- name: GetTasksGroupedByHeadings :many
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
ORDER BY h.id;

-- name: GetTasksForToday :many
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
ORDER BY l.id;

-- name: GetUpcomingTasks :many
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
          (t.start_date >= COALESCE(@after_date::timestamptz, CURRENT_DATE + interval '1 day'))
              AND (t.deleted_at IS NULL)
              AND (COALESCE(t.start_date, @after_date::timestamptz) > @after_date::timestamptz)
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
LIMIT $2;

-- name: GetOverdueTasks :many
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
          AND (t.deleted_at IS NULL OR l.id > @after_id::varchar)
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
LIMIT $2;

-- name: GetTasksForSomeday :many
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
          AND (t.deleted_at IS NULL OR l.id > @after_id::varchar)
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
LIMIT $2;

-- name: GetCompletedTasks :many
SELECT
    DATE_TRUNC('month', t.updated_at)::timestamptz AS month,
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
      WHERE statuses.title = @status_title::varchar
      )
      AND (t.deleted_at IS NULL
               OR (DATE_TRUNC('month', t.updated_at) > @after_date::timestamptz AND t.deleted_at IS NULL))
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
        tags,
        t.updated_at
    ) t
GROUP BY month
ORDER BY month
LIMIT $2;

-- name: GetArchivedTasks :many
SELECT
    DATE_TRUNC('month', t.updated_at)::timestamptz AS month,
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
        WHERE statuses.title = @status_title::varchar
        )
      AND (t.deleted_at IS NOT NULL
               OR (DATE_TRUNC('month', t.updated_at) > @after_month::timestamptz AND t.deleted_at IS NOT NULL)
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
        tags,
        t.updated_at,
        t.deleted_at
    ) t
GROUP BY month
ORDER BY month DESC
LIMIT $2;

-- name: MoveTaskToAnotherList :one
UPDATE tasks
SET	list_id = $1,
    heading_id = $2,
    updated_at = $3
WHERE id = $4
  AND user_id = $5
  AND deleted_at IS NULL
RETURNING id;

-- name: MoveTaskToAnotherHeading :one
UPDATE tasks
SET	heading_id = $1,
    updated_at = $2
WHERE id = $3
    AND user_id = $4
    AND deleted_at IS NULL
RETURNING id;

-- name: MarkTaskAsCompleted :one
UPDATE tasks
SET	status_id = $1,
    updated_at = $2
WHERE id = $3
  AND user_id = $4
  AND deleted_at IS NULL
RETURNING id;

-- name: MarkTaskAsArchived :one
UPDATE tasks
SET status_id = $1, deleted_at = $2
WHERE id = $3
  AND user_id = $4
  AND deleted_at IS NULL
RETURNING id;