-- name: CreateTag :exec
INSERT INTO tags (id, title, user_id, updated_at)
VALUES ($1, LOWER($2), $3, $4);

-- name: LinkTagToTask :exec
INSERT INTO tasks_tags (task_id, tag_id)
VALUES ($1, (SELECT id
             FROM tags
             WHERE title = LOWER($2))
);

-- name: UnlinkTagFromTask :exec
DELETE FROM tasks_tags
WHERE task_id = $1
  AND tag_id = (SELECT id
                FROM tags
                WHERE title = LOWER($2)
);

-- name: GetTagIDByTitle :one
SELECT id
FROM tags
WHERE title = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: GetTagsByUserID :many
SELECT id, title, updated_at
FROM tags
WHERE user_id = $1
  AND deleted_at IS NULL;

-- name: GetTagsByTaskID :many
SELECT tags.id, tags.title, tags.updated_at
FROM tags
    JOIN tasks_tags
        ON tags.id = tasks_tags.tag_id
WHERE tasks_tags.task_id = $1
  AND tags.deleted_at IS NULL;

