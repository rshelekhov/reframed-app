-- name: GetStatuses :many
SELECT id, title
FROM statuses;

-- name: GetStatusID :one
SELECT id
FROM statuses
WHERE title = $1;

-- name: GetStatusName :one
SELECT title
FROM statuses
WHERE id = $1;