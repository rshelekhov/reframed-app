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
}

func NewTaskHandler(
	log logger.Interface,
	tokenAuth *jwtoken.JWTAuth,
	taskStorage storage.TaskStorage,
) *TaskHandler {
	return &TaskHandler{
		Logger:      log,
		TokenAuth:   tokenAuth,
		TaskStorage: taskStorage,
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
			UserID:      userID,
		}

		err = h.TaskStorage.CreateTask(r.Context(), newTask)
		if err != nil {
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
			UserID:      userID,
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
