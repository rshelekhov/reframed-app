-- name: CreateList :exec
INSERT INTO lists (id, title, user_id, is_default, created_at,updated_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetListByID :one
SELECT id, title, user_id, is_default, updated_at
FROM lists
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: GetListsByUserID :many
SELECT id, title, updated_at
FROM lists
WHERE user_id = $1
  AND deleted_at IS NULL
ORDER BY id;

-- name: GetDefaultListID :one
SELECT id
FROM lists
WHERE user_id = $1
  AND is_default = TRUE
  AND deleted_at IS NULL;

-- name: UpdateList :one
UPDATE lists
SET title = $1,	updated_at = $2
WHERE id = $3
  AND user_id = $4
  AND deleted_at IS NULL
RETURNING id;

-- name: DeleteList :one
UPDATE lists
SET deleted_at = $1
WHERE id = $2
  AND user_id = $3
  AND deleted_at IS NULL
RETURNING id;