// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: heading.sql

package sqlc

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

const createHeading = `-- name: CreateHeading :exec
INSERT INTO headings (id, title, list_id, user_id, is_default, created_at,updated_at)
VALUES($1, $2, $3, $4, $5, $6, $7)
`

type CreateHeadingParams struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	ListID    string    `db:"list_id"`
	UserID    string    `db:"user_id"`
	IsDefault bool      `db:"is_default"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (q *Queries) CreateHeading(ctx context.Context, arg CreateHeadingParams) error {
	_, err := q.db.Exec(ctx, createHeading,
		arg.ID,
		arg.Title,
		arg.ListID,
		arg.UserID,
		arg.IsDefault,
		arg.CreatedAt,
		arg.UpdatedAt,
	)
	return err
}

const deleteHeading = `-- name: DeleteHeading :one
UPDATE headings
SET deleted_at = $1
WHERE id = $2
  AND user_id = $3
  AND deleted_at IS NULL
RETURNING id
`

type DeleteHeadingParams struct {
	DeletedAt pgtype.Timestamptz `db:"deleted_at"`
	ID        string             `db:"id"`
	UserID    string             `db:"user_id"`
}

func (q *Queries) DeleteHeading(ctx context.Context, arg DeleteHeadingParams) (string, error) {
	row := q.db.QueryRow(ctx, deleteHeading, arg.DeletedAt, arg.ID, arg.UserID)
	var id string
	err := row.Scan(&id)
	return id, err
}

const deleteHeadingsByListID = `-- name: DeleteHeadingsByListID :exec
UPDATE headings
SET deleted_at = $1
WHERE list_id = $2
  AND user_id = $3
  AND deleted_at IS NULL
`

type DeleteHeadingsByListIDParams struct {
	DeletedAt pgtype.Timestamptz `db:"deleted_at"`
	ListID    string             `db:"list_id"`
	UserID    string             `db:"user_id"`
}

func (q *Queries) DeleteHeadingsByListID(ctx context.Context, arg DeleteHeadingsByListIDParams) error {
	_, err := q.db.Exec(ctx, deleteHeadingsByListID, arg.DeletedAt, arg.ListID, arg.UserID)
	return err
}

const getDefaultHeadingID = `-- name: GetDefaultHeadingID :one
SELECT id
FROM headings
WHERE list_id = $1
  AND user_id = $2
  AND is_default = TRUE
  AND deleted_at IS NULL
`

type GetDefaultHeadingIDParams struct {
	ListID string `db:"list_id"`
	UserID string `db:"user_id"`
}

func (q *Queries) GetDefaultHeadingID(ctx context.Context, arg GetDefaultHeadingIDParams) (string, error) {
	row := q.db.QueryRow(ctx, getDefaultHeadingID, arg.ListID, arg.UserID)
	var id string
	err := row.Scan(&id)
	return id, err
}

const getHeadingByID = `-- name: GetHeadingByID :one
SELECT id, title, list_id, user_id, updated_at
FROM headings
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type GetHeadingByIDParams struct {
	ID     string `db:"id"`
	UserID string `db:"user_id"`
}

type GetHeadingByIDRow struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	ListID    string    `db:"list_id"`
	UserID    string    `db:"user_id"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (q *Queries) GetHeadingByID(ctx context.Context, arg GetHeadingByIDParams) (GetHeadingByIDRow, error) {
	row := q.db.QueryRow(ctx, getHeadingByID, arg.ID, arg.UserID)
	var i GetHeadingByIDRow
	err := row.Scan(
		&i.ID,
		&i.Title,
		&i.ListID,
		&i.UserID,
		&i.UpdatedAt,
	)
	return i, err
}

const getHeadingsByListID = `-- name: GetHeadingsByListID :many
SELECT id, title, list_id, user_id, updated_at
FROM headings
WHERE list_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type GetHeadingsByListIDParams struct {
	ListID string `db:"list_id"`
	UserID string `db:"user_id"`
}

type GetHeadingsByListIDRow struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	ListID    string    `db:"list_id"`
	UserID    string    `db:"user_id"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (q *Queries) GetHeadingsByListID(ctx context.Context, arg GetHeadingsByListIDParams) ([]GetHeadingsByListIDRow, error) {
	rows, err := q.db.Query(ctx, getHeadingsByListID, arg.ListID, arg.UserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetHeadingsByListIDRow{}
	for rows.Next() {
		var i GetHeadingsByListIDRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.ListID,
			&i.UserID,
			&i.UpdatedAt,
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

const moveHeadingToAnotherList = `-- name: MoveHeadingToAnotherList :one
UPDATE headings
SET list_id = $1, updated_at = $2
WHERE id = $3
  AND user_id = $4
  AND deleted_at IS NULL
RETURNING id
`

type MoveHeadingToAnotherListParams struct {
	ListID    string    `db:"list_id"`
	UpdatedAt time.Time `db:"updated_at"`
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
}

func (q *Queries) MoveHeadingToAnotherList(ctx context.Context, arg MoveHeadingToAnotherListParams) (string, error) {
	row := q.db.QueryRow(ctx, moveHeadingToAnotherList,
		arg.ListID,
		arg.UpdatedAt,
		arg.ID,
		arg.UserID,
	)
	var id string
	err := row.Scan(&id)
	return id, err
}

const updateHeading = `-- name: UpdateHeading :one
UPDATE headings
SET title = $1, updated_at = $2
WHERE id = $3
  AND user_id = $4
  AND deleted_at IS NULL
RETURNING id
`

type UpdateHeadingParams struct {
	Title     string    `db:"title"`
	UpdatedAt time.Time `db:"updated_at"`
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
}

func (q *Queries) UpdateHeading(ctx context.Context, arg UpdateHeadingParams) (string, error) {
	row := q.db.QueryRow(ctx, updateHeading,
		arg.Title,
		arg.UpdatedAt,
		arg.ID,
		arg.UserID,
	)
	var id string
	err := row.Scan(&id)
	return id, err
}

const updateTasksListID = `-- name: UpdateTasksListID :exec
UPDATE tasks
SET list_id = $1, updated_at = $2
WHERE heading_id = $3
  AND user_id = $4
`

type UpdateTasksListIDParams struct {
	ListID    string    `db:"list_id"`
	UpdatedAt time.Time `db:"updated_at"`
	HeadingID string    `db:"heading_id"`
	UserID    string    `db:"user_id"`
}

func (q *Queries) UpdateTasksListID(ctx context.Context, arg UpdateTasksListIDParams) error {
	_, err := q.db.Exec(ctx, updateTasksListID,
		arg.ListID,
		arg.UpdatedAt,
		arg.HeadingID,
		arg.UserID,
	)
	return err
}
