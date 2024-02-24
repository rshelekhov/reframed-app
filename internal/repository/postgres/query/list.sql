-- name: CreateList :exec
INSERT INTO lists (id, title, user_id, updated_at)
VALUES ($1, $2, $3, $4);

-- name: GetListByID :one
SELECT id, title, user_id, updated_at
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

-- name: UpdateList :exec
UPDATE lists
SET title = $1,	updated_at = $2
WHERE id = $3
  AND user_id = $4;

-- name: DeleteList :exec
UPDATE lists
SET deleted_at = $1
WHERE id = $2
  AND user_id = $3;