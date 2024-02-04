package handlers

import (
	"errors"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/logger"
	"github.com/rshelekhov/reframed/src/models"
	"github.com/rshelekhov/reframed/src/server/middleware/jwtoken"
	"github.com/rshelekhov/reframed/src/storage"
	"github.com/segmentio/ksuid"
	"log/slog"
	"net/http"
)

type TaskHandler struct {
	Logger      logger.Interface
	TokenAuth   *jwtoken.JWTAuth
	TaskStorage storage.TaskStorage
	TagStorage  storage.TagStorage
}

func NewTaskHandler(
	log logger.Interface,
	tokenAuth *jwtoken.JWTAuth,
	taskStorage storage.TaskStorage,
	tagStorage storage.TagStorage,
) *TaskHandler {
	return &TaskHandler{
		Logger:      log,
		TokenAuth:   tokenAuth,
		TaskStorage: taskStorage,
		TagStorage:  tagStorage,
	}
}

func (h *TaskHandler) CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.CreateTask"

		log := logger.LogWithRequest(h.Logger, op, r)

		task := &models.Task{}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID, err := GetIDFromQuery(w, r, log, c.ListIDKey)
		if err != nil {
			return
		}

		err = DecodeAndValidateJSON(w, r, log, task)
		if err != nil {
			return
		}

		newTask := models.Task{
			ID:          ksuid.New().String(),
			Title:       task.Title,
			Description: task.Description,
			StartDate:   task.StartDate,
			Deadline:    task.Deadline,
			StartTime:   task.StartTime,
			EndTime:     task.EndTime,
			ListID:      listID,
			Tags:        task.Tags,
			UserID:      userID,
		}

		for _, tag := range task.Tags {
			if err = h.TagStorage.CreateTagIfNotExists(r.Context(), tag, userID); err != nil {
				handleInternalServerError(w, r, log, c.ErrFailedToCreateTag, err)
				return
			}
		}

		if err = h.TagStorage.LinkTagsToTask(r.Context(), newTask.ID, task.Tags); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToLinkTagsToTask, err)
			return
		}

		if err = h.TaskStorage.CreateTask(r.Context(), newTask); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToCreateTask, err)
			return
		}

		handleResponseSuccess(w, r, log, "task created", newTask, slog.String(c.TaskIDKey, newTask.ID))
	}
}

func (h *TaskHandler) GetTaskByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetTaskByID"

		log := logger.LogWithRequest(h.Logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID, err := GetIDFromQuery(w, r, log, c.TaskIDKey)
		if err != nil {
			return
		}

		task, err := h.TaskStorage.GetTaskByID(r.Context(), taskID, userID)
		switch {
		case errors.Is(err, c.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task received", task, slog.String(c.TaskIDKey, task.ID))
		}
	}
}

func (h *TaskHandler) GetTasksByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetTasksByUserID"

		log := logger.LogWithRequest(h.Logger, op, r)

		pagination, err := ParseLimitAndOffset(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrFailedToParsePagination, err)
			return
		}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		tasks, err := h.TaskStorage.GetTasksByUserID(r.Context(), userID, pagination)
		switch {
		case errors.Is(err, c.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks found", tasks,
				slog.Int(c.CountKey, len(tasks)),
				slog.Int(c.LimitKey, pagination.Limit),
				slog.Int(c.OffsetKey, pagination.Offset),
			)
		}
	}
}

func (h *TaskHandler) GetTasksByListID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetTasksByListID"

		log := logger.LogWithRequest(h.Logger, op, r)

		pagination, err := ParseLimitAndOffset(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrFailedToParsePagination, err)
			return
		}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID, err := GetIDFromQuery(w, r, log, c.ListIDKey)
		if err != nil {
			return
		}

		tasks, err := h.TaskStorage.GetTasksByListID(r.Context(), listID, userID, pagination)
		switch {
		case errors.Is(err, c.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks found", tasks,
				slog.Int(c.CountKey, len(tasks)),
				slog.Int(c.LimitKey, pagination.Limit),
				slog.Int(c.OffsetKey, pagination.Offset),
			)
		}
	}
}

func (h *TaskHandler) UpdateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.UpdateTask"

		log := logger.LogWithRequest(h.Logger, op, r)
		task := &models.UpdateTask{}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID, err := GetIDFromQuery(w, r, log, c.TaskIDKey)
		if err != nil {
			return
		}

		if err = DecodeAndValidateJSON(w, r, log, task); err != nil {
			return
		}

		updatedTask := models.Task{
			ID:          taskID,
			Title:       task.Title,
			Description: task.Description,
			StartDate:   task.StartDate,
			Deadline:    task.Deadline,
			StartTime:   task.StartTime,
			EndTime:     task.EndTime,
			Tags:        task.Tags,
			UserID:      userID,
		}

		currentTags, err := h.TagStorage.GetTagsByTaskID(r.Context(), taskID)
		if errors.Is(err, c.ErrNoTagsFound) {
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTagsFound)
			return
		}
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		}

		tagsToAdd, tagsToRemove := findTagsToAddAndRemove(currentTags, updatedTask.Tags)

		for _, tag := range tagsToAdd {
			if err = h.TagStorage.CreateTagIfNotExists(r.Context(), tag, userID); err != nil {
				handleInternalServerError(w, r, log, c.ErrFailedToCreateTag, err)
				return
			}
		}

		if err = h.TagStorage.LinkTagsToTask(r.Context(), updatedTask.ID, tagsToAdd); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToLinkTagsToTask, err)
			return
		}

		if err = h.TagStorage.UnlinkTagsFromTask(r.Context(), updatedTask.ID, tagsToRemove); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToDeleteTag, err)
			return
		}

		err = h.TaskStorage.UpdateTask(r.Context(), updatedTask)
		switch {
		case errors.Is(err, c.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToUpdateTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task updated", updatedTask, slog.String(c.TaskIDKey, taskID))
		}
	}
}

func findTagsToAddAndRemove(currentTags, updatedTags []string) (tagsToAdd, tagsToRemove []string) {
	tagMap := make(map[string]bool)

	for _, tag := range currentTags {
		tagMap[tag] = true
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

func (h *TaskHandler) UpdateTaskTime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.UpdateTaskTimes"

		log := logger.LogWithRequest(h.Logger, op, r)
		task := &models.UpdateTask{}

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID, err := GetIDFromQuery(w, r, log, c.TaskIDKey)
		if err != nil {
			return
		}

		if err = DecodeAndValidateJSON(w, r, log, task); err != nil {
			return
		}

		updatedTask := models.Task{
			ID:        taskID,
			StartTime: task.StartTime,
			EndTime:   task.EndTime,
			UserID:    userID,
		}

		err = h.TaskStorage.UpdateTask(r.Context(), updatedTask)
		switch {
		case errors.Is(err, c.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToUpdateTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task updated", updatedTask, slog.String(c.TaskIDKey, taskID))
		}
	}
}

func (h *TaskHandler) CompleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.CompleteTask"

		log := logger.LogWithRequest(h.Logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID, err := GetIDFromQuery(w, r, log, c.TaskIDKey)
		if err != nil {
			return
		}

		err = h.TaskStorage.CompleteTask(r.Context(), taskID, userID)
		switch {
		case errors.Is(err, c.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToCompleteTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task completed", nil, slog.String(c.TaskIDKey, taskID))
		}
	}
}

func (h *TaskHandler) DeleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.DeleteTask"

		log := logger.LogWithRequest(h.Logger, op, r)

		_, claims, err := jwtoken.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID, err := GetIDFromQuery(w, r, log, c.TaskIDKey)
		if err != nil {
			return
		}

		err = h.TaskStorage.DeleteTask(r.Context(), taskID, userID)
		switch {
		case errors.Is(err, c.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToDeleteTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task deleted", models.Task{ID: taskID}, slog.String(c.TaskIDKey, taskID))
		}
	}
}
