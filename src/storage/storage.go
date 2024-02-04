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
		GetUser(ctx context.Context, userID string) (models.User, error)
		GetUsers(ctx context.Context, pgn models.Pagination) ([]models.User, error)
		UpdateUser(ctx context.Context, user models.User) error
		DeleteUser(ctx context.Context, id string) error
	}

	// ListStorage defines the list repository
	ListStorage interface {
		CreateList(ctx context.Context, list models.List) error
		GetListByID(ctx context.Context, listID, userID string) (models.List, error)
		GetLists(ctx context.Context, userID string, pgn models.Pagination) ([]models.List, error)
		UpdateList(ctx context.Context, list models.List) error
		DeleteList(ctx context.Context, listID, userID string) error
	}

	// TaskStorage defines the task repository
	TaskStorage interface {
		CreateTask(ctx context.Context, task models.Task) error
		GetTaskByID(ctx context.Context, taskID, userID string) (models.Task, error)
		GetTasksByUserID(ctx context.Context, userID string, pgn models.Pagination) ([]models.Task, error)
		GetTasksByListID(ctx context.Context, listID, userID string, pgn models.Pagination) ([]models.Task, error)
		UpdateTask(ctx context.Context, task models.Task) error
		UpdateTaskTime(ctx context.Context, task models.Task) error
		CompleteTask(ctx context.Context, taskID, userID string) error
		DeleteTask(ctx context.Context, taskID, userID string) error
	}

	// TagStorage defines the tag repository
	TagStorage interface {
		CreateTagIfNotExists(ctx context.Context, tag, userID string) error
		LinkTagsToTask(ctx context.Context, taskID string, tags []string) error
		UnlinkTagsFromTask(ctx context.Context, taskID string, tagsToRemove []string) error
		GetTagsByUserID(ctx context.Context, userID string) ([]models.Tag, error)
		GetTagsByTaskID(ctx context.Context, taskID string) ([]string, error)
	}
)
