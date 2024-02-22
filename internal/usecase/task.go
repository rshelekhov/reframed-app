package usecase

import (
	"context"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
	"github.com/rshelekhov/reframed/pkg/constants/le"
	"github.com/segmentio/ksuid"
	"time"
)

type TaskUsecase struct {
	taskStorage    port.TaskStorage
	headingUsecase port.HeadingUsecase
	tagUsecase     port.TagUsecase
	txManager      port.TransactionManager
}

func NewTaskUsecase(
	storage port.TaskStorage,
	headingUsecase port.HeadingUsecase,
	tagUsecase port.TagUsecase,
	txManager port.TransactionManager,
) *TaskUsecase {
	return &TaskUsecase{
		taskStorage:    storage,
		headingUsecase: headingUsecase,
		tagUsecase:     tagUsecase,
		txManager:      txManager,
	}
}

func (u *TaskUsecase) CreateTask(ctx context.Context, data *model.TaskRequestData) (string, error) {
	const op = "task.usecase.CreateTask"

	if data.HeadingID == "" {
		defaultHeadingID, err := u.headingUsecase.GetDefaultHeadingID(ctx, model.HeadingRequestData{
			ListID: data.ListID,
			UserID: data.UserID,
		})
		if err != nil {
			return "", err
		}
		data.HeadingID = defaultHeadingID
	}

	statusNotStarted, err := u.taskStorage.GetTaskStatusID(ctx, model.StatusNotStarted)
	if err != nil {
		return "", err
	}
	data.StatusID = statusNotStarted

	newTask := model.Task{
		ID:          ksuid.New().String(),
		Title:       data.Title,
		Description: data.Description,
		StartDate:   data.StartDate,
		Deadline:    data.Deadline,
		StartTime:   data.StartTime,
		EndTime:     data.EndTime,
		StatusID:    data.StatusID,
		ListID:      data.ListID,
		HeadingID:   data.HeadingID,
		UserID:      data.UserID,
		UpdatedAt:   time.Now(),
	}

	err = u.txManager.WithinTransaction(ctx, op, func(ctx context.Context) error {
		return u.createTaskWithinTransaction(ctx, newTask)
	})
	if err != nil {
		return "", err
	}

	return newTask.ID, nil
}

func (u *TaskUsecase) createTaskWithinTransaction(ctx context.Context, task model.Task) error {
	for _, tag := range task.Tags {
		if err := u.tagUsecase.CreateTagIfNotExists(ctx, model.TagRequestData{
			Title:  tag,
			UserID: task.UserID,
		}); err != nil {
			return err
		}
	}

	if err := u.taskStorage.CreateTask(ctx, task); err != nil {
		return err
	}

	if err := u.tagUsecase.LinkTagsToTask(ctx, task.ID, task.Tags); err != nil {
		return err
	}

	return nil
}

func (u *TaskUsecase) GetTaskByID(ctx context.Context, data model.TaskRequestData) (model.TaskResponseData, error) {
	task, err := u.taskStorage.GetTaskByID(ctx, data.ID, data.UserID)
	if err != nil {
		return model.TaskResponseData{}, err
	}
	return model.TaskResponseData{
		ID:        task.ID,
		Title:     task.Title,
		StartDate: task.StartDate,
		Deadline:  task.Deadline,
		StartTime: task.StartTime,
		EndTime:   task.EndTime,
		StatusID:  task.StatusID,
		ListID:    task.ListID,
		HeadingID: task.HeadingID,
		UserID:    task.UserID,
		UpdatedAt: task.UpdatedAt,
	}, nil
}

func (u *TaskUsecase) GetTasksByUserID(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskResponseData, error) {
	tasks, err := u.taskStorage.GetTasksByUserID(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}

	var tasksResp []model.TaskResponseData
	for _, task := range tasks {
		tasksResp = append(tasksResp, mapTaskToResponseData(task))
	}
	return tasksResp, nil
}

func (u *TaskUsecase) GetTasksByListID(ctx context.Context, data model.TaskRequestData) ([]model.TaskResponseData, error) {
	tasks, err := u.taskStorage.GetTasksByListID(ctx, data.ListID, data.UserID)
	if err != nil {
		return nil, err
	}

	var tasksResp []model.TaskResponseData
	for _, task := range tasks {
		tasksResp = append(tasksResp, mapTaskToResponseData(task))
	}
	return tasksResp, nil
}

func mapTaskToResponseData(task model.Task) model.TaskResponseData {
	return model.TaskResponseData{
		ID:        task.ID,
		Title:     task.Title,
		StartDate: task.StartDate,
		Deadline:  task.Deadline,
		StartTime: task.StartTime,
		EndTime:   task.EndTime,
		StatusID:  task.StatusID,
		ListID:    task.ListID,
		HeadingID: task.HeadingID,
		UserID:    task.UserID,
		UpdatedAt: task.UpdatedAt,
	}
}

func (u *TaskUsecase) GetTasksGroupedByHeadings(ctx context.Context, data model.TaskRequestData) ([]model.TaskGroup, error) {
	taskGroups, err := u.taskStorage.GetTasksGroupedByHeadings(ctx, data.ListID, data.UserID)
	if err != nil {
		return nil, err
	}
	return taskGroups, nil
}

func (u *TaskUsecase) GetTasksForToday(ctx context.Context, userID string) ([]model.TaskGroup, error) {
	taskGroups, err := u.taskStorage.GetTasksForToday(ctx, userID)
	if err != nil {
		return nil, err
	}
	return taskGroups, nil
}

func (u *TaskUsecase) GetUpcomingTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	taskGroups, err := u.taskStorage.GetUpcomingTasks(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}
	return taskGroups, nil
}

func (u *TaskUsecase) GetOverdueTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	taskGroups, err := u.taskStorage.GetOverdueTasks(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}
	return taskGroups, nil
}

func (u *TaskUsecase) GetTasksForSomeday(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	taskGroups, err := u.taskStorage.GetTasksForSomeday(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}
	return taskGroups, nil
}

func (u *TaskUsecase) GetCompletedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	taskGroups, err := u.taskStorage.GetCompletedTasks(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}
	return taskGroups, nil
}

func (u *TaskUsecase) GetArchivedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	taskGroups, err := u.taskStorage.GetArchivedTasks(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}
	return taskGroups, nil
}

func (u *TaskUsecase) UpdateTask(ctx context.Context, data *model.TaskRequestData) error {
	const op = "task.usecase.UpdateTask"

	updatedTask := model.Task{
		ID:        data.ID,
		Title:     data.Title,
		StartDate: data.StartDate,
		Deadline:  data.Deadline,
		StartTime: data.StartTime,
		EndTime:   data.EndTime,
		ListID:    data.ListID,
		HeadingID: data.HeadingID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	err := u.txManager.WithinTransaction(ctx, op, func(ctx context.Context) error {
		return u.updateTaskWithinTransaction(ctx, updatedTask)
	})
	if err != nil {
		return err
	}

	return nil
}

func (u *TaskUsecase) updateTaskWithinTransaction(ctx context.Context, task model.Task) error {
	currentTags, err := u.tagUsecase.GetTagsByTaskID(ctx, task.ID)
	if err != nil {
		return err
	}

	// Use transactions in these methods
	tagsToAdd, tagsToRemove := findTagsToAddAndRemove(currentTags, task.Tags)

	for _, tag := range task.Tags {
		if err = u.tagUsecase.CreateTagIfNotExists(ctx, model.TagRequestData{
			Title:  tag,
			UserID: task.UserID,
		}); err != nil {
			return err
		}
	}

	if err = u.taskStorage.UpdateTask(ctx, task); err != nil {
		return err
	}

	if err = u.tagUsecase.UnlinkTagsFromTask(ctx, task.ID, tagsToRemove); err != nil {
		return err
	}

	if err = u.tagUsecase.LinkTagsToTask(ctx, task.ID, tagsToAdd); err != nil {
		return err
	}

	return nil
}

func findTagsToAddAndRemove(currentTags []model.TagResponseData, updatedTags []string) (tagsToAdd, tagsToRemove []string) {
	tagMap := make(map[string]bool)

	for _, tag := range currentTags {
		tagMap[tag.Title] = true
	}

	for _, tag := range updatedTags {
		if _, ok := tagMap[tag]; ok {
			delete(tagMap, tag)
		} else {
			tagsToAdd = append(tagsToAdd, tag)
		}
	}

	for tag := range tagMap {
		tagsToRemove = append(tagsToRemove, tag)
	}

	return tagsToAdd, tagsToRemove
}

func (u *TaskUsecase) UpdateTaskTime(ctx context.Context, data *model.TaskRequestData) error {
	if !data.StartTime.IsZero() && !data.EndTime.IsZero() {
		taskStatusID, err := u.taskStorage.GetTaskStatusID(ctx, model.StatusPlanned)
		if err != nil {
			return err
		}
		data.StatusID = taskStatusID
	} else if data.StartTime.IsZero() && data.EndTime.IsZero() {
		taskStatusID, err := u.taskStorage.GetTaskStatusID(ctx, model.StatusNotStarted)
		if err != nil {
			return err
		}
		data.StatusID = taskStatusID
	} else {
		return le.ErrInvalidTaskTimeRange
	}

	updatedTask := model.Task{
		ID:        data.ID,
		StartTime: data.StartTime,
		EndTime:   data.EndTime,
		StatusID:  data.StatusID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	return u.taskStorage.UpdateTaskTime(ctx, updatedTask)
}

func (u *TaskUsecase) MoveTaskToAnotherList(ctx context.Context, data model.TaskRequestData) error {
	defaultHeadingID, err := u.headingUsecase.GetDefaultHeadingID(ctx, model.HeadingRequestData{
		ListID: data.ListID,
		UserID: data.UserID,
	})
	if err != nil {
		return err
	}
	data.HeadingID = defaultHeadingID

	updatedTask := model.Task{
		ID:        data.ID,
		ListID:    data.ListID,
		HeadingID: data.HeadingID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	return u.taskStorage.MoveTaskToAnotherList(ctx, updatedTask)
}

func (u *TaskUsecase) CompleteTask(ctx context.Context, data model.TaskRequestData) error {
	statusCompleted, err := u.taskStorage.GetTaskStatusID(ctx, model.StatusCompleted)
	if err != nil {
		return err
	}
	data.StatusID = statusCompleted

	completedTask := model.Task{
		ID:        data.ID,
		StatusID:  data.StatusID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	return u.taskStorage.MarkAsCompleted(ctx, completedTask)
}

func (u *TaskUsecase) ArchiveTask(ctx context.Context, data model.TaskRequestData) error {
	statusArchived, err := u.taskStorage.GetTaskStatusID(ctx, model.StatusArchived)
	if err != nil {
		return err
	}
	data.StatusID = statusArchived

	archivedTask := model.Task{
		ID:        data.ID,
		StatusID:  data.StatusID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	return u.taskStorage.MarkAsArchived(ctx, archivedTask)
}
