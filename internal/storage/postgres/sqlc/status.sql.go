// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: status.sql

package sqlc

import (
	"context"
)

const getStatusByID = `-- name: GetStatusByID :one
SELECT title
FROM statuses
WHERE id = $1
`

func (q *Queries) GetStatusByID(ctx context.Context, id int32) (string, error) {
	row := q.db.QueryRow(ctx, getStatusByID, id)
	var title string
	err := row.Scan(&title)
	return title, err
}

const getStatuses = `-- name: GetStatuses :many
SELECT id, title
FROM statuses
`

func (q *Queries) GetStatuses(ctx context.Context) ([]Status, error) {
	rows, err := q.db.Query(ctx, getStatuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Status{}
	for rows.Next() {
		var i Status
		if err := rows.Scan(&i.ID, &i.Title); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
