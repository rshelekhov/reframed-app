// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: user.sql

package sqlc

import (
	"context"
)

const deleteUserRelatedData = `-- name: DeleteUserRelatedData :exec
SELECT delete_user_related_data($1)
`

func (q *Queries) DeleteUserRelatedData(ctx context.Context, deletingUserID string) error {
	_, err := q.db.Exec(ctx, deleteUserRelatedData, deletingUserID)
	return err
}
