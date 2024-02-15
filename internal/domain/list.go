package domain

import (
	"context"
	"time"
)

// List DB domain
type (
	List struct {
		ID        string     `db:"id"`
		Title     string     `db:"headingTitle"`
		UserID    string     `db:"user_id"`
		UpdatedAt *time.Time `db:"updated_at"`
		DeletedAt *time.Time `db:"deleted_at"`
	}

	ListRequestData struct {
		ID     string `json:"id"`
		Title  string `json:"title" validate:"required"`
		UserID string `json:"user_id"`
	}

	ListResponseData struct {
		ID        string     `json:"id"`
		Title     string     `json:"title"`
		UserID    string     `json:"user_id"`
		UpdatedAt *time.Time `json:"updated_at"`
	}
)

type listTitle string

func (t listTitle) String() string {
	return string(t)
}

const (
	DefaultInboxList listTitle = "Inbox"
)

type (
	ListUsecase interface {
		CreateList(ctx context.Context, data *ListRequestData) (string, error)
		CreateDefaultList(ctx context.Context, list List) error
		GetListByID(ctx context.Context, data ListRequestData) (ListResponseData, error)
		GetListsByUserID(ctx context.Context, userID string) ([]ListResponseData, error)
		UpdateList(ctx context.Context, data *ListRequestData) error
		DeleteList(ctx context.Context, data ListRequestData) error
	}

	ListStorage interface {
		CreateList(ctx context.Context, list List) error
		GetListByID(ctx context.Context, listID, userID string) (List, error)
		GetListsByUserID(ctx context.Context, userID string) ([]List, error)
		UpdateList(ctx context.Context, list List) error
		DeleteList(ctx context.Context, list List) error
	}
)
