--
-- SQL queries for user management
--

-- name: GetUserStatus :one
SELECT CASE
WHEN EXISTS(
    SELECT 1
    FROM users
    WHERE users.email = $1
      AND deleted_at IS NULL FOR UPDATE
) THEN 'active'
WHEN EXISTS(
    SELECT 1
    FROM users
    WHERE users.email = $1
      AND deleted_at IS NOT NULL FOR UPDATE
    ) THEN 'soft_deleted'
ELSE 'not_found' END AS status;

-- name: SetDeletedUserAtNull :exec
UPDATE users
SET deleted_at = NULL
WHERE email = $1;

-- name: InsertUser :exec
INSERT INTO users (id, email, password_hash, updated_at)
VALUES ($1, $2, $3, $4);

-- name: UpdateLatestLoginAt :exec
UPDATE user_devices
SET latest_login_at = $1
WHERE id = $2;

-- name: GetUserByEmail :one
SELECT id, email, updated_at
FROM users
WHERE email = $1
  AND deleted_at IS NULL;

-- name: GetUserByID :one
SELECT id, email, updated_at
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetUserData :one
SELECT id, email, password_hash, updated_at
FROM users
WHERE id = $1
  AND deleted_at IS NULL;

-- name: GetUserID :one
SELECT id
FROM users
WHERE email = $1
  AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = $1
WHERE id = $2
  AND deleted_at IS NULL;

--
-- SQL queries for user sessions
--

-- name: AddDevice :exec
INSERT INTO user_devices (id, user_id, user_agent, ip, detached, latest_login_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetUserDeviceID :one
SELECT id
FROM user_devices
WHERE user_id = $1
  AND user_agent = $2
  AND detached = FALSE;

-- name: SaveSession :exec
INSERT INTO refresh_sessions (user_id, device_id, refresh_token, last_visit_at, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetSessionByRefreshToken :one
SELECT user_id, device_id, last_visit_at, expires_at
FROM refresh_sessions
WHERE refresh_token = $1;

-- name: DeleteRefreshTokenFromSession :exec
DELETE FROM refresh_sessions
WHERE refresh_token = $1;

-- name: DeleteSession :exec
DELETE FROM refresh_sessions
WHERE user_id = $1
  AND device_id = $2;