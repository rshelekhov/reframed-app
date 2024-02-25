-- name: CreateHeading :exec
INSERT INTO headings (id, title, list_id, user_id, is_default, updated_at)
VALUES($1, $2, $3, $4, $5, $6);

-- name: GetDefaultHeadingID :one
SELECT id
FROM headings
WHERE list_id = $1
  AND user_id = $2
  AND is_default = TRUE
  AND deleted_at IS NULL;

-- name: GetHeadingByID :one
SELECT id, title, list_id, user_id, updated_at
FROM headings
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: GetHeadingsByListID :many
SELECT id, title, list_id, user_id, updated_at
FROM headings
WHERE list_id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: UpdateHeading :exec
UPDATE headings
SET title = $1, updated_at = $2
WHERE id = $3
  AND user_id = $4;

-- name: MoveHeadingToAnotherList :exec
UPDATE headings
SET list_id = $1, updated_at = $2
WHERE id = $3
  AND user_id = $4;

-- name: UpdateTasksListID :exec
UPDATE tasks
SET list_id = $1, updated_at = $2
WHERE heading_id = $3
  AND user_id = $4;

-- name: DeleteHeading :exec
UPDATE headings
SET deleted_at = $1
WHERE id = $2
  AND user_id = $3
  AND deleted_at IS NULL;