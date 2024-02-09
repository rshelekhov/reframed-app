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
		RemoveSession(ctx context.Context, userID, deviceID string) error
		GetSessionByRefreshToken(ctx context.Context, refreshToken string) (models.Session, error)
		AddDevice(ctx context.Context, device models.UserDevice) error
		GetUserDevice(ctx context.Context, userID, userAgent string) (models.UserDevice, error)
		GetUserByEmail(ctx context.Context, email string) (models.User, error)
		GetUserByID(ctx context.Context, id string) (models.User, error)
		GetUserProfile(ctx context.Context, userID string) (models.User, error)
		UpdateUser(ctx context.Context, user models.User) error
		DeleteUser(ctx context.Context, id string) error
	}

	// ListStorage defines the list repository
	ListStorage interface {
		CreateList(ctx context.Context, list models.List) error
		GetListByID(ctx context.Context, listID, userID string) (models.List, error)
		GetListsByUserID(ctx context.Context, userID string) ([]models.List, error)
		UpdateList(ctx context.Context, list models.List) error
		DeleteList(ctx context.Context, listID, userID string) error
	}

	// TaskStorage defines the task repository
	TaskStorage interface {
		CreateTask(ctx context.Context, task models.Task) error
		GetTaskByID(ctx context.Context, taskID, userID string) (models.Task, error)
		GetTasksByUserID(ctx context.Context, userID string, pgn models.Pagination) ([]models.Task, error)
		GetTasksByListID(ctx context.Context, listID, userID string) ([]models.Task, error)
		GetTasksGroupedByHeadings(ctx context.Context, listID, userID string) ([]models.TaskGroup, error)
		GetTasksForToday(ctx context.Context, userID string) ([]models.TaskGroup, error)
		GetUpcomingTasks(ctx context.Context, userID string, pgn models.Pagination) ([]models.TaskGroup, error)
		GetOverdueTasks(ctx context.Context, userID string, pgn models.Pagination) ([]models.TaskGroup, error)
		GetTasksForSomeday(ctx context.Context, userID string, pgn models.Pagination) ([]models.TaskGroup, error)
		GetCompletedTasks(ctx context.Context, userID string, pgn models.Pagination) ([]models.TaskGroup, error)
		GetArchivedTasks(ctx context.Context, userID string, pgn models.Pagination) ([]models.TaskGroup, error)
		UpdateTask(ctx context.Context, task models.Task) error
		UpdateTaskTime(ctx context.Context, task models.Task) error
		MoveTaskToAnotherList(ctx context.Context, listID, taskID, userID string) error
		CompleteTask(ctx context.Context, taskID, userID string) error
		ArchiveTask(ctx context.Context, taskID, userID string) error
	}

	// HeadingStorage defines the heading repository
	HeadingStorage interface {
		CreateHeading(ctx context.Context, heading models.Heading, isDefault bool) error
		GetDefaultHeadingID(ctx context.Context, listID, userID string) (string, error)
		GetHeadingByID(ctx context.Context, headingID, userID string) (models.Heading, error)
		GetHeadingsByListID(ctx context.Context, listID, userID string) ([]models.Heading, error)
		UpdateHeading(ctx context.Context, heading models.Heading) error
		MoveHeadingToAnotherList(ctx context.Context, headingID, otherListID, userID string) error
		DeleteHeading(ctx context.Context, headingID, userID string) error
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
