package port

import (
	"context"

	"github.com/rshelekhov/reframed/internal/model"
)

type (
	HeadingUsecase interface {
		CreateHeading(ctx context.Context, data *model.HeadingRequestData) (model.HeadingResponseData, error)
		CreateDefaultHeading(ctx context.Context, heading model.Heading) error
		GetHeadingByID(ctx context.Context, data model.HeadingRequestData) (model.HeadingResponseData, error)
		GetDefaultHeadingID(ctx context.Context, data model.HeadingRequestData) (string, error)
		GetHeadingsByListID(ctx context.Context, data model.HeadingRequestData) ([]model.HeadingResponseData, error)
		UpdateHeading(ctx context.Context, data *model.HeadingRequestData) (model.HeadingResponseData, error)
		MoveHeadingToAnotherList(ctx context.Context, data *model.HeadingRequestData) (model.HeadingResponseData, error)
		DeleteHeading(ctx context.Context, data *model.HeadingRequestData) error
	}

	HeadingStorage interface {
		CreateHeading(ctx context.Context, heading model.Heading) error
		GetDefaultHeadingID(ctx context.Context, listID, userID string) (string, error)
		GetHeadingByID(ctx context.Context, headingID, userID string) (model.Heading, error)
		GetHeadingsByListID(ctx context.Context, listID, userID string) ([]model.Heading, error)
		UpdateHeading(ctx context.Context, heading model.Heading) error
		MoveHeadingToAnotherList(ctx context.Context, heading model.Heading, task model.Task) error
		DeleteHeading(ctx context.Context, heading model.Heading) error
	}
)
