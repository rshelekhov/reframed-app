package port

import "context"

type (
	UserUsecase interface {
		DeleteUserRelatedData(ctx context.Context, userID string) error
	}

	UserStorage interface {
		DeleteUserData(ctx context.Context, userID string) error
	}
)
