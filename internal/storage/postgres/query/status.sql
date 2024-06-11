-- name: GetStatuses :many
SELECT id, title
FROM statuses;

-- name: GetStatusByID :one
SELECT title
FROM statuses
WHERE id = $1;