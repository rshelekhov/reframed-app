package storage

import (
	"context"
	"github.com/rshelekhov/reframed/src/models"
)

type (
	// UserStorage defines the user repository
	UserStorage interface {
		CreateUser(ctx context.Context, user models.User) error
		SaveSession(ctx context.Context, userID, deviceID string, session models.Session) error
		GetSessionByRefreshToken(ctx context.Context, refreshToken string) (models.Session, error)
		AddDevice(ctx context.Context, device models.UserDevice) error
		GetUserDevice(ctx context.Context, userID, userAgent string) (models.UserDevice, error)
		GetUserCredentials(ctx context.Context, user *models.User) (models.User, error)
		GetUser(ctx context.Context, id string) (models.User, error)
		GetUsers(ctx context.Context, pgn models.Pagination) ([]models.User, error)
		UpdateUser(ctx context.Context, user models.User) error
		DeleteUser(ctx context.Context, id string) error
	}

	// ListStorage defines the list repository
	ListStorage interface {
		CreateList(ctx context.Context, list models.List) error
		GetListByID(ctx context.Context, id string) (models.List, error)
		GetLists(ctx context.Context, userID string, pgn models.Pagination) ([]models.List, error)
		UpdateList(ctx context.Context, list models.List) error
		DeleteList(ctx context.Context, id string) error
	}
)
