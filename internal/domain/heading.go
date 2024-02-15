package domain

import (
	"context"
	"time"
)

// Heading DB domain
type (
	Heading struct {
		ID        string     `db:"id"`
		Title     string     `db:"headingTitle"`
		ListID    string     `db:"list_id"`
		UserID    string     `db:"user_id"`
		IsDefault bool       `db:"is_default"`
		UpdatedAt *time.Time `db:"updated_at"`
		DeletedAt *time.Time `db:"deleted_at"`
	}

	HeadingRequestData struct {
		ID     string `json:"id"`
		Title  string `json:"title" validate:"required"`
		ListID string `json:"list_id"`
		UserID string `json:"user_id"`
	}

	HeadingResponseData struct {
		ID        string     `json:"id,omitempty"`
		Title     string     `json:"title,omitempty"`
		ListID    string     `json:"list_id,omitempty"`
		UserID    string     `json:"user_id,omitempty"`
		UpdatedAt *time.Time `json:"updated_at,omitempty"`
	}
)

type headingTitle string

func (t headingTitle) String() string {
	return string(t)
}

const (
	DefaultHeading headingTitle = "Default"
)

type (
	HeadingUsecase interface {
		CreateHeading(ctx context.Context, data *HeadingRequestData) (string, error)
		CreateDefaultHeading(ctx context.Context, heading Heading) error
		GetHeadingByID(ctx context.Context, data HeadingRequestData) (HeadingResponseData, error)
		GetDefaultHeadingID(ctx context.Context, data HeadingRequestData) (string, error)
		GetHeadingsByListID(ctx context.Context, data HeadingRequestData) ([]HeadingResponseData, error)
		UpdateHeading(ctx context.Context, data *HeadingRequestData) error
		MoveHeadingToAnotherList(ctx context.Context, data HeadingRequestData) error
		DeleteHeading(ctx context.Context, data HeadingRequestData) error
	}

	HeadingStorage interface {
		CreateHeading(ctx context.Context, heading Heading) error
		GetDefaultHeadingID(ctx context.Context, listID, userID string) (string, error)
		GetHeadingByID(ctx context.Context, headingID, userID string) (Heading, error)
		GetHeadingsByListID(ctx context.Context, listID, userID string) ([]Heading, error)
		UpdateHeading(ctx context.Context, heading Heading) error
		MoveHeadingToAnotherList(ctx context.Context, heading Heading, task Task) error
		DeleteHeading(ctx context.Context, heading Heading) error
	}
)
