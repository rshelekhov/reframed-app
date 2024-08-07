package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/go-chi/chi/v5"

	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type taskHandler struct {
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.TaskUsecase
}

func newTaskHandler(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.TaskUsecase,
) *taskHandler {
	return &taskHandler{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (h *taskHandler) CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.CreateTask"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		listID := chi.URLParam(r, key.ListID)
		headingID := chi.URLParam(r, key.HeadingID)

		taskInput := &model.TaskRequestData{}
		if err = decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.HeadingID = headingID
		taskInput.ListID = listID
		taskInput.UserID = userID

		taskResponse, err := h.usecase.CreateTask(ctx, taskInput)
		switch {
		case errors.Is(err, le.ErrDefaultListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrDefaultListNotFound)
		case errors.Is(err, le.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToCreateTask, err)
		default:
			handleResponseCreated(w, r, log, "task created", taskResponse, slog.String(key.TaskID, taskResponse.ID))
		}
	}
}

func (h *taskHandler) CreateTaskInDefaultList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.CreateTask"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		taskInput := &model.TaskRequestData{}
		if err = decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.UserID = userID

		taskResponse, err := h.usecase.CreateTask(ctx, taskInput)
		switch {
		case errors.Is(err, le.ErrDefaultListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrDefaultListNotFound)
		case errors.Is(err, le.ErrDefaultHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToCreateTask, err)
		default:
			handleResponseCreated(w, r, log, "task created", taskResponse, slog.String(key.TaskID, taskResponse.ID))
		}
	}
}

func (h *taskHandler) GetTaskByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetTaskByID"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		taskID := chi.URLParam(r, key.TaskID)

		taskInput := model.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		taskResp, err := h.usecase.GetTaskByID(ctx, taskInput)
		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "task received", taskResp, slog.String(key.TaskID, taskResp.ID))
		}
	}
}

func (h *taskHandler) GetTasksByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetTasksByUserID"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		pagination, err := ParseLimitAndCursor(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidCursor)
		}

		tasksResp, err := h.usecase.GetTasksByUserID(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks found", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "tasks found", tasksResp)
		}
	}
}

func (h *taskHandler) GetTasksByListID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetTasksByListID"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		listID := chi.URLParam(r, key.ListID)

		tasksInput := model.TaskRequestData{
			ListID: listID,
			UserID: userID,
		}

		tasksResp, err := h.usecase.GetTasksByListID(ctx, tasksInput)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks found for the list", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "tasks for the list found", tasksResp)
		}
	}
}

func (h *taskHandler) GetTasksGroupedByHeadings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetTasksGroupedByHeading"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		listID := chi.URLParam(r, key.ListID)

		tasksInput := model.TaskRequestData{
			ListID: listID,
			UserID: userID,
		}

		tasksResp, err := h.usecase.GetTasksGroupedByHeading(ctx, tasksInput)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks grouped by headings found", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "tasks grouped by headings found", tasksResp)
		}
	}
}

func (h *taskHandler) GetTasksForToday() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetTasksForToday"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		tasksResp, err := h.usecase.GetTasksForToday(ctx, userID)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks found for today", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "tasks for today found", tasksResp)
		}
	}
}

func (h *taskHandler) GetUpcomingTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetUpcomingTasks"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		pagination, err := ParseLimitAndCursor(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidCursor)
		}

		tasksResp, err := h.usecase.GetUpcomingTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no upcoming tasks found", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "upcoming tasks found", tasksResp)
		}
	}
}

func (h *taskHandler) GetOverdueTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetOverdueTasks"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		pagination, err := ParseLimitAndCursor(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidCursor)
		}

		tasksResp, err := h.usecase.GetOverdueTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no overdue tasks found", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "overdue tasks found", tasksResp)
		}
	}
}

func (h *taskHandler) GetTasksForSomeday() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetTasksForSomeday"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		pagination, err := ParseLimitAndCursor(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidCursor)
		}

		tasksResp, err := h.usecase.GetTasksForSomeday(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks for someday found", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "tasks for someday found", tasksResp)
		}
	}
}

func (h *taskHandler) GetCompletedTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetCompletedTasks"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		pagination, err := ParseLimitAndCursor(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidCursor)
		}

		tasksResp, err := h.usecase.GetCompletedTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no completed tasks found", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "completed tasks found", tasksResp)
		}
	}
}

func (h *taskHandler) GetArchivedTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.GetArchivedTasks"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		pagination, err := ParseLimitAndCursor(r)
		if err != nil {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidCursor)
		}

		tasksResp, err := h.usecase.GetArchivedTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no archived tasks found", nil)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "archived tasks found", tasksResp)
		}
	}
}

func (h *taskHandler) UpdateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.UpdateTask"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		taskID := chi.URLParam(r, key.TaskID)

		taskInput := &model.TaskRequestData{}
		if err = decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.ID = taskID
		taskInput.UserID = userID

		taskResponse, err := h.usecase.UpdateTask(ctx, taskInput)
		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateTask, err)
		default:
			handleResponseSuccess(w, r, log, "task updated", taskResponse, slog.String(key.TaskID, taskResponse.ID))
		}
	}
}

func (h *taskHandler) UpdateTaskTime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.UpdateTaskTimes"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		taskID := chi.URLParam(r, key.TaskID)

		taskInput := &model.TaskRequestTimeData{}
		if err = decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.ID = taskID
		taskInput.UserID = userID

		taskResponse, err := h.usecase.UpdateTaskTime(ctx, taskInput)
		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
		case errors.Is(err, le.ErrInvalidTaskTimeRange):
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidTaskTimeRange)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateTask, err)
		default:
			handleResponseSuccess(w, r, log, "task updated", taskResponse, slog.String(key.TaskID, taskResponse.ID))
		}
	}
}

func (h *taskHandler) MoveTaskToAnotherList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.MoveTaskToAnotherList"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		taskID := chi.URLParam(r, key.TaskID)

		listID := r.URL.Query().Get(key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryListID)
		}

		taskInput := model.TaskRequestData{
			ID:     taskID,
			ListID: listID,
			UserID: userID,
		}

		taskResponse, err := h.usecase.MoveTaskToAnotherList(ctx, taskInput)
		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
		case errors.Is(err, le.ErrListNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrListNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToMoveTask, err)
		default:
			handleResponseSuccess(w, r, log, "task moved to another list", taskResponse, slog.String(key.TaskID, taskInput.ID))
		}
	}
}

func (h *taskHandler) MoveTaskToAnotherHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.MoveTaskToAnotherHeading"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		taskID := chi.URLParam(r, key.TaskID)

		headingID := r.URL.Query().Get(key.HeadingID)
		if headingID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryHeadingID)
		}

		taskInput := model.TaskRequestData{
			ID:        taskID,
			HeadingID: headingID,
			UserID:    userID,
		}

		taskResponse, err := h.usecase.MoveTaskToAnotherHeading(ctx, taskInput)
		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
		case errors.Is(err, le.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToMoveTask, err)
		default:
			handleResponseSuccess(w, r, log, "task moved to another heading", taskResponse, slog.String(key.TaskID, taskInput.ID))
		}
	}
}

func (h *taskHandler) CompleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.CompleteTask"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		taskID := chi.URLParam(r, key.TaskID)

		taskInput := model.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		taskResponse, err := h.usecase.CompleteTask(ctx, taskInput)
		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToCompleteTask, err)
		default:
			handleResponseSuccess(w, r, log, "task completed", taskResponse, slog.String(key.TaskID, taskID))
		}
	}
}

func (h *taskHandler) ArchiveTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handler.ArchiveTask"

		ctx := r.Context()
		log := logger.LogWithRequest(h.logger, op, r)

		userID, err := getUserIDFromContext(ctx, w, r, h.jwt, log)
		if err != nil {
			return
		}

		taskID := chi.URLParam(r, key.TaskID)

		taskInput := model.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		taskResponse, err := h.usecase.ArchiveTask(ctx, taskInput)
		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToArchiveTask, err)
		default:
			handleResponseSuccess(w, r, log, "task archived", taskResponse, slog.String(key.TaskID, taskID))
		}
	}
}
