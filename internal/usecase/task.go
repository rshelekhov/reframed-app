package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/ksuid"

	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type TaskUsecase struct {
	storage        port.TaskStorage
	HeadingUsecase port.HeadingUsecase
	TagUsecase     port.TagUsecase
	ListUsecase    port.ListUsecase
}

func NewTaskUsecase(storage port.TaskStorage) *TaskUsecase {
	return &TaskUsecase{storage: storage}
}

func (u *TaskUsecase) CreateTask(ctx context.Context, data *model.TaskRequestData) (model.TaskResponseData, error) {
	if data.ListID == "" {
		defaultListID, err := u.ListUsecase.GetDefaultListID(ctx, data.UserID)
		if err != nil {
			return model.TaskResponseData{}, err
		}

		data.ListID = defaultListID
	}

	if data.HeadingID == "" {
		defaultHeadingID, err := u.HeadingUsecase.GetDefaultHeadingID(ctx, model.HeadingRequestData{
			ListID: data.ListID,
			UserID: data.UserID,
		})
		if err != nil {
			return model.TaskResponseData{}, err
		}

		data.HeadingID = defaultHeadingID
	}

	statusNotStarted, err := u.storage.GetTaskStatusID(ctx, model.StatusNotStarted)
	if err != nil {
		return model.TaskResponseData{}, err
	}

	data.StatusID = statusNotStarted

	newTask := model.Task{
		ID:          ksuid.New().String(),
		Title:       data.Title,
		Description: data.Description,
		StartDate:   data.StartDateParsed,
		Deadline:    data.DeadlineParsed,
		StartTime:   data.StartTimeParsed,
		EndTime:     data.EndTimeParsed,
		StatusID:    data.StatusID,
		ListID:      data.ListID,
		HeadingID:   data.HeadingID,
		UserID:      data.UserID,
		Tags:        data.Tags,
		UpdatedAt:   time.Now(),
	}

	if err = u.storage.Transaction(ctx, func(_ port.TaskStorage) error {
		for _, tag := range newTask.Tags {
			if err = u.TagUsecase.CreateTagIfNotExists(ctx, model.TagRequestData{
				Title:  tag,
				UserID: newTask.UserID,
			}); err != nil {
				return err
			}
		}
		if err = u.storage.CreateTask(ctx, newTask); err != nil {
			return err
		}
		if err = u.TagUsecase.LinkTagsToTask(ctx, newTask.UserID, newTask.ID, newTask.Tags); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.TaskResponseData{}, err
	}

	return model.TaskResponseData{
		ID:          newTask.ID,
		Title:       newTask.Title,
		Description: newTask.Description,
		StartDate:   newTask.StartDate,
		Deadline:    newTask.Deadline,
		StartTime:   newTask.StartTime,
		EndTime:     newTask.EndTime,
		StatusID:    newTask.StatusID,
		ListID:      newTask.ListID,
		HeadingID:   newTask.HeadingID,
		UserID:      newTask.UserID,
		Tags:        newTask.Tags,
		UpdatedAt:   newTask.UpdatedAt,
	}, nil
}

func (u *TaskUsecase) GetTaskByID(ctx context.Context, data model.TaskRequestData) (model.TaskResponseData, error) {
	task, err := u.storage.GetTaskByID(ctx, data.ID, data.UserID)
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
		Tags:      task.Tags,
		UpdatedAt: task.UpdatedAt,
	}, nil
}

func (u *TaskUsecase) GetTasksByUserID(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskResponseData, error) {
	tasks, err := u.storage.GetTasksByUserID(ctx, userID, pgn)
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
	tasks, err := u.storage.GetTasksByListID(ctx, data.ListID, data.UserID)
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
		Tags:      task.Tags,
		UpdatedAt: task.UpdatedAt,
	}
}

func (u *TaskUsecase) GetTasksGroupedByHeading(ctx context.Context, data model.TaskRequestData) ([]model.TaskGroupWithHeading, error) {
	const op = "task.usecase.GetTasksGroupedByHeading"

	groupsRaw, err := u.storage.GetTasksGroupedByHeadings(ctx, data.ListID, data.UserID)
	if err != nil {
		return nil, err
	}

	var taskGroups []model.TaskGroupWithHeading

	for _, group := range groupsRaw {
		var taskGroup model.TaskGroupWithHeading

		var tasks []model.TaskResponseData

		err = json.Unmarshal(group.Tasks, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks from postgres json object: %w", op, err)
		}

		taskGroup.HeadingID = group.HeadingID
		taskGroup.Tasks = tasks

		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, nil
}

func (u *TaskUsecase) GetTasksForToday(ctx context.Context, userID string) ([]model.TaskGroup, error) {
	const op = "task.usecase.GetTasksForToday"

	groupsRaw, err := u.storage.GetTasksForToday(ctx, userID)
	if err != nil {
		return nil, err
	}

	var taskGroups []model.TaskGroup

	for _, group := range groupsRaw {
		var taskGroup model.TaskGroup

		var tasks []model.TaskResponseData

		err = json.Unmarshal(group.Tasks, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks from postgres json object: %w", op, err)
		}

		taskGroup.ListID = group.ListID
		taskGroup.Tasks = tasks

		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, nil
}

func (u *TaskUsecase) GetUpcomingTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.UpcomingTaskGroup, error) {
	const op = "task.usecase.GetUpcomingTasks"

	groupsRaw, err := u.storage.GetUpcomingTasks(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}

	var taskGroups []model.UpcomingTaskGroup

	for _, group := range groupsRaw {
		var taskGroup model.UpcomingTaskGroup

		var tasks []model.TaskResponseData

		err = json.Unmarshal(group.Tasks, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks from postgres json object: %w", op, err)
		}

		taskGroup.StartDate = group.StartDate
		taskGroup.Tasks = tasks

		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, nil
}

func (u *TaskUsecase) GetOverdueTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroup, error) {
	const op = "task.usecase.GetOverdueTasks"

	groupsRaw, err := u.storage.GetOverdueTasks(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}

	var taskGroups []model.TaskGroup

	for _, group := range groupsRaw {
		var taskGroup model.TaskGroup

		var tasks []model.TaskResponseData

		err = json.Unmarshal(group.Tasks, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks from postgres json object: %w", op, err)
		}

		taskGroup.ListID = group.ListID
		taskGroup.Tasks = tasks

		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, nil
}

func (u *TaskUsecase) GetTasksForSomeday(ctx context.Context, userID string, pgn model.Pagination) ([]model.TaskGroupForSomeday, error) {
	const op = "task.usecase.GetTasksForSomeday"

	groupsRaw, err := u.storage.GetTasksForSomeday(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}

	var taskGroups []model.TaskGroupForSomeday

	for _, group := range groupsRaw {
		var taskGroup model.TaskGroupForSomeday

		var tasks []model.TaskResponseData

		err = json.Unmarshal(group.Tasks, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks from postgres json object: %w", op, err)
		}

		taskGroup.ListID = group.ListID
		taskGroup.Tasks = tasks

		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, nil
}

func (u *TaskUsecase) GetCompletedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.CompletedTasksGroup, error) {
	const op = "task.usecase.GetCompletedTasks"

	groupsRaw, err := u.storage.GetCompletedTasks(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}

	var taskGroups []model.CompletedTasksGroup

	for _, group := range groupsRaw {
		var taskGroup model.CompletedTasksGroup

		var tasks []model.TaskResponseData

		err = json.Unmarshal(group.Tasks, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks from postgres json object: %w", op, err)
		}

		taskGroup.Month = group.Month
		taskGroup.Tasks = tasks

		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, nil
}

func (u *TaskUsecase) GetArchivedTasks(ctx context.Context, userID string, pgn model.Pagination) ([]model.ArchivedTasksGroup, error) {
	const op = "task.usecase.GetArchivedTasks"

	groupsRaw, err := u.storage.GetArchivedTasks(ctx, userID, pgn)
	if err != nil {
		return nil, err
	}

	var taskGroups []model.ArchivedTasksGroup

	for _, group := range groupsRaw {
		var taskGroup model.ArchivedTasksGroup

		var tasks []model.TaskResponseData

		err = json.Unmarshal(group.Tasks, &tasks)
		if err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal tasks from postgres json object: %w", op, err)
		}

		taskGroup.Month = group.Month
		taskGroup.Tasks = tasks

		taskGroups = append(taskGroups, taskGroup)
	}

	return taskGroups, nil
}

func (u *TaskUsecase) UpdateTask(ctx context.Context, data *model.TaskRequestData) (model.TaskResponseData, error) {
	updatedTask := model.Task{
		ID:        data.ID,
		Title:     data.Title,
		StartDate: data.StartDateParsed,
		Deadline:  data.DeadlineParsed,
		StartTime: data.StartTimeParsed,
		EndTime:   data.EndTimeParsed,
		ListID:    data.ListID,
		HeadingID: data.HeadingID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err := u.storage.Transaction(ctx, func(_ port.TaskStorage) error {
		currentTags, err := u.TagUsecase.GetTagsByTaskID(ctx, updatedTask.ID)
		if err != nil {
			return err
		}

		tagsToAdd, tagsToRemove := findTagsToAddAndRemove(currentTags, updatedTask.Tags)

		for _, tag := range updatedTask.Tags {
			if err = u.TagUsecase.CreateTagIfNotExists(ctx, model.TagRequestData{
				Title:  tag,
				UserID: updatedTask.UserID,
			}); err != nil {
				return err
			}
		}

		if err = u.storage.UpdateTask(ctx, updatedTask); err != nil {
			return err
		}
		if err = u.TagUsecase.UnlinkTagsFromTask(ctx, updatedTask.UserID, updatedTask.ID, tagsToRemove); err != nil {
			return err
		}
		if err = u.TagUsecase.LinkTagsToTask(ctx, updatedTask.UserID, updatedTask.ID, tagsToAdd); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.TaskResponseData{}, err
	}

	return model.TaskResponseData{
		ID:        updatedTask.ID,
		Title:     updatedTask.Title,
		StartDate: updatedTask.StartDate,
		Deadline:  updatedTask.Deadline,
		StartTime: updatedTask.StartTime,
		EndTime:   updatedTask.EndTime,
		StatusID:  updatedTask.StatusID,
		ListID:    updatedTask.ListID,
		HeadingID: updatedTask.HeadingID,
		UserID:    updatedTask.UserID,
		Tags:      updatedTask.Tags,
		UpdatedAt: updatedTask.UpdatedAt,
	}, nil
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

func (u *TaskUsecase) UpdateTaskTime(ctx context.Context, data *model.TaskRequestTimeData) (model.TaskResponseTimeData, error) {
	var statusID int

	switch {
	case !data.StartTimeParsed.IsZero() && !data.EndTimeParsed.IsZero():
		taskStatusID, err := u.storage.GetTaskStatusID(ctx, model.StatusPlanned)
		if err != nil {
			return model.TaskResponseTimeData{}, err
		}
		statusID = taskStatusID
	case data.StartTimeParsed.IsZero() && data.EndTimeParsed.IsZero():
		taskStatusID, err := u.storage.GetTaskStatusID(ctx, model.StatusNotStarted)
		if err != nil {
			return model.TaskResponseTimeData{}, err
		}
		statusID = taskStatusID
	default:
		return model.TaskResponseTimeData{}, le.ErrInvalidTaskTimeRange
	}

	updatedTaskTime := model.Task{
		ID:        data.ID,
		StartTime: data.StartTimeParsed,
		EndTime:   data.EndTimeParsed,
		StatusID:  statusID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err := u.storage.UpdateTaskTime(ctx, updatedTaskTime); err != nil {
		return model.TaskResponseTimeData{}, err
	}

	return model.TaskResponseTimeData{
		ID:        updatedTaskTime.ID,
		StartTime: updatedTaskTime.StartTime,
		EndTime:   updatedTaskTime.EndTime,
		UserID:    updatedTaskTime.UserID,
		UpdatedAt: updatedTaskTime.UpdatedAt,
	}, nil
}

func (u *TaskUsecase) MoveTaskToAnotherList(ctx context.Context, data model.TaskRequestData) (model.TaskResponseData, error) {
	// Check if list exists
	_, err := u.ListUsecase.GetListByID(ctx, model.ListRequestData{
		ID:     data.ListID,
		UserID: data.UserID,
	})
	if err != nil {
		return model.TaskResponseData{}, err
	}

	defaultHeadingID, err := u.HeadingUsecase.GetDefaultHeadingID(ctx, model.HeadingRequestData{
		ListID: data.ListID,
		UserID: data.UserID,
	})
	if err != nil {
		return model.TaskResponseData{}, err
	}

	updatedTask := model.Task{
		ID:        data.ID,
		ListID:    data.ListID,
		HeadingID: defaultHeadingID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err = u.storage.MoveTaskToAnotherList(ctx, updatedTask); err != nil {
		return model.TaskResponseData{}, err
	}

	return model.TaskResponseData{
		ID:        updatedTask.ID,
		ListID:    updatedTask.ListID,
		HeadingID: updatedTask.HeadingID,
		UserID:    updatedTask.UserID,
		UpdatedAt: updatedTask.UpdatedAt,
	}, nil
}

func (u *TaskUsecase) MoveTaskToAnotherHeading(ctx context.Context, data model.TaskRequestData) (model.TaskResponseData, error) {
	// Check if heading exists
	_, err := u.HeadingUsecase.GetHeadingByID(ctx, model.HeadingRequestData{
		ID:     data.HeadingID,
		UserID: data.UserID,
	})
	if err != nil {
		return model.TaskResponseData{}, err
	}

	updatedTask := model.Task{
		ID:        data.ID,
		HeadingID: data.HeadingID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err = u.storage.MoveTaskToAnotherHeading(ctx, updatedTask); err != nil {
		return model.TaskResponseData{}, err
	}

	return model.TaskResponseData{
		ID:        updatedTask.ID,
		HeadingID: updatedTask.HeadingID,
		UserID:    updatedTask.UserID,
		UpdatedAt: updatedTask.UpdatedAt,
	}, nil
}

func (u *TaskUsecase) CompleteTask(ctx context.Context, data model.TaskRequestData) (model.TaskResponseData, error) {
	statusCompleted, err := u.storage.GetTaskStatusID(ctx, model.StatusCompleted)
	if err != nil {
		return model.TaskResponseData{}, err
	}

	data.StatusID = statusCompleted

	completedTask := model.Task{
		ID:        data.ID,
		StatusID:  data.StatusID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err = u.storage.MarkAsCompleted(ctx, completedTask); err != nil {
		return model.TaskResponseData{}, err
	}

	return model.TaskResponseData{
		ID:        completedTask.ID,
		StatusID:  completedTask.StatusID,
		UserID:    completedTask.UserID,
		UpdatedAt: completedTask.UpdatedAt,
	}, nil

}

func (u *TaskUsecase) ArchiveTask(ctx context.Context, data model.TaskRequestData) (model.TaskResponseData, error) {
	statusArchived, err := u.storage.GetTaskStatusID(ctx, model.StatusArchived)
	if err != nil {
		return model.TaskResponseData{}, err
	}

	data.StatusID = statusArchived

	archivedTask := model.Task{
		ID:        data.ID,
		StatusID:  data.StatusID,
		UserID:    data.UserID,
		UpdatedAt: time.Now(),
	}

	if err = u.storage.MarkAsArchived(ctx, archivedTask); err != nil {
		return model.TaskResponseData{}, err
	}

	return model.TaskResponseData{
		ID:        archivedTask.ID,
		StatusID:  archivedTask.StatusID,
		UserID:    archivedTask.UserID,
		UpdatedAt: archivedTask.UpdatedAt,
	}, nil
}
