package port

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
)

type (
	TaskUsecase interface {
		CreateTask(ctx context.Context, data *model.TaskRequestData) (model.TaskResponseData, error)
		GetTaskByID(ctx context.Context, data model.TaskRequestData) (model.TaskResponseData, error)
		GetTasksByUserID(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskResponseData, error)
		GetTasksByListID(ctx context.Context, data model.TaskRequestData) ([]model.TaskResponseData, error)
		GetTasksGroupedByHeadings(ctx context.Context, data model.TaskRequestData) ([]model.TaskGroup, error)
		GetTasksForToday(ctx context.Context, userID string) ([]model.TaskGroup, error)
		GetUpcomingTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		GetOverdueTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		GetTasksForSomeday(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		GetCompletedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		GetArchivedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		UpdateTask(ctx context.Context, data *model.TaskRequestData) (model.TaskResponseData, error)
		UpdateTaskTime(ctx context.Context, data *model.TaskRequestTimeData) (model.TaskResponseTimeData, error)
		MoveTaskToAnotherList(ctx context.Context, data model.TaskRequestData) error
		CompleteTask(ctx context.Context, data model.TaskRequestData) error
		ArchiveTask(ctx context.Context, data model.TaskRequestData) error
	}

	TaskStorage interface {
		Transaction(ctx context.Context, fn func(storage TaskStorage) error) error
		CreateTask(ctx context.Context, task model.Task) error
		GetTaskStatusID(ctx context.Context, status model.StatusName) (int, error)
		GetTaskByID(ctx context.Context, taskID, userID string) (model.Task, error)
		GetTasksByUserID(ctx context.Context, userID string, pgn model.Pagination) ([]model.Task, error)
		GetTasksByListID(ctx context.Context, listID, userID string) ([]model.Task, error)
		GetTasksGroupedByHeadings(ctx context.Context, listID, userID string) ([]model.TaskGroup, error)
		GetTasksForToday(ctx context.Context, userID string) ([]model.TaskGroup, error)
		GetUpcomingTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		GetOverdueTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		GetTasksForSomeday(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		GetCompletedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		GetArchivedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error)
		UpdateTask(ctx context.Context, task model.Task) error
		UpdateTaskTime(ctx context.Context, task model.Task) error
		MoveTaskToAnotherList(ctx context.Context, task model.Task) error
		MarkAsCompleted(ctx context.Context, task model.Task) error
		MarkAsArchived(ctx context.Context, task model.Task) error
	}
)
