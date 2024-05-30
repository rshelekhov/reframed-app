package v1

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/go-chi/chi/v5"

	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/rshelekhov/reframed/internal/port"
)

type taskController struct {
	logger  *slog.Logger
	jwt     *jwtoken.TokenService
	usecase port.TaskUsecase
}

func newTaskController(
	log *slog.Logger,
	jwt *jwtoken.TokenService,
	usecase port.TaskUsecase,
) *taskController {
	return &taskController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}
}

func (c *taskController) CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.CreateTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryListID)
			return
		}

		headingID := chi.URLParam(r, key.HeadingID)

		taskInput := &model.TaskRequestData{}
		if err = decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.HeadingID = headingID
		taskInput.ListID = listID
		taskInput.UserID = userID

		taskResponse, err := c.usecase.CreateTask(ctx, taskInput)

		switch {
		case errors.Is(err, le.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToCreateTask, err)
			return
		default:
			handleResponseCreated(w, r, log, "task created", taskResponse, slog.String(key.TaskID, taskResponse.ID))
		}
	}
}

func (c *taskController) CreateTaskInDefaultList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.CreateTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		taskInput := &model.TaskRequestData{}
		if err = decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.UserID = userID

		taskResponse, err := c.usecase.CreateTask(ctx, taskInput)

		switch {
		case errors.Is(err, le.ErrDefaultHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToCreateTask, err)
			return
		default:
			handleResponseCreated(w, r, log, "task created", taskResponse, slog.String(key.TaskID, taskResponse.ID))
		}
	}
}

func (c *taskController) GetTaskByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTaskByID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryTaskID)
			return
		}

		taskInput := model.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		taskResp, err := c.usecase.GetTaskByID(ctx, taskInput)

		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task received", taskResp, slog.String(key.TaskID, taskResp.ID))
		}
	}
}

func (c *taskController) GetTasksByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTasksByUserID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		pagination := ParseLimitAndAfterID(r)

		tasksResp, err := c.usecase.GetTasksByUserID(ctx, userID, pagination)

		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks found", tasksResp,
				slog.Int(key.Count, len(tasksResp)),
			)
		}
	}
}

func (c *taskController) GetTasksByListID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTasksByListID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryListID)
			return
		}

		tasksInput := model.TaskRequestData{
			ListID: listID,
			UserID: userID,
		}

		tasksResp, err := c.usecase.GetTasksByListID(ctx, tasksInput)

		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks found", tasksResp,
				slog.Int(key.Count, len(tasksResp)),
			)
		}
	}
}

func (c *taskController) GetTasksGroupedByHeadings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTasksGroupedByHeadings"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryListID)
			return
		}

		tasksInput := model.TaskRequestData{
			ListID: listID,
			UserID: userID,
		}

		tasksResp, err := c.usecase.GetTasksGroupedByHeadings(ctx, tasksInput)

		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks grouped by headings found", tasksResp, slog.Int(key.Count, len(tasksResp)))
		}
	}
}

func (c *taskController) GetTasksForToday() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTasksForToday"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		tasksResp, err := c.usecase.GetTasksForToday(ctx, userID)

		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks for today found", tasksResp, slog.Int(key.Count, len(tasksResp)))
		}
	}
}

func (c *taskController) GetUpcomingTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetUpcomingTasks"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToParseQueryParams, err)
			return
		}

		tasksResp, err := c.usecase.GetUpcomingTasks(ctx, userID, pagination)

		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "upcoming tasks found", tasksResp, slog.Int(key.Count, len(tasksResp)))
		}
	}
}

func (c *taskController) GetOverdueTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetOverdueTasks"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		pagination := ParseLimitAndAfterID(r)

		tasksResp, err := c.usecase.GetOverdueTasks(ctx, userID, pagination)

		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "overdue tasks found", tasksResp, slog.Int(key.Count, len(tasksResp)))
		}
	}
}

func (c *taskController) GetTasksForSomeday() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTasksForSomeday"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		pagination := ParseLimitAndAfterID(r)

		tasksResp, err := c.usecase.GetTasksForSomeday(ctx, userID, pagination)

		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks for someday found", tasksResp, slog.Int(key.Count, len(tasksResp)))
		}
	}
}

func (c *taskController) GetCompletedTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetCompletedTasks"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToParseQueryParams, err)
			return
		}

		tasksResp, err := c.usecase.GetCompletedTasks(ctx, userID, pagination)

		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "completed tasks found", tasksResp, slog.Int(key.Count, len(tasksResp)))
		}
	}
}

func (c *taskController) GetArchivedTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetArchivedTasks"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToParseQueryParams, err)
			return
		}

		tasksResp, err := c.usecase.GetArchivedTasks(ctx, userID, pagination)

		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "archived tasks found", tasksResp, slog.Int(key.Count, len(tasksResp)))
		}
	}
}

func (c *taskController) UpdateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.UpdateTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryTaskID)
			return
		}

		taskInput := &model.TaskRequestData{}
		if err = decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.ID = taskID
		taskInput.UserID = userID

		taskResponse, err := c.usecase.UpdateTask(ctx, taskInput)

		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task updated", taskResponse, slog.String(key.TaskID, taskResponse.ID))
		}
	}
}

func (c *taskController) UpdateTaskTime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.UpdateTaskTimes"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryTaskID)
			return
		}

		taskInput := &model.TaskRequestTimeData{}
		if err = decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.ID = taskID
		taskInput.UserID = userID

		taskResponse, err := c.usecase.UpdateTaskTime(ctx, taskInput)

		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task updated", taskResponse, slog.String(key.TaskID, taskResponse.ID))
		}
	}
}

func (c *taskController) MoveTaskToAnotherList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.MoveTaskToAnotherList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryTaskID)
			return
		}

		listID := r.URL.Query().Get(key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryListID)
			return
		}

		taskInput := model.TaskRequestData{
			ID:     taskID,
			ListID: listID,
			UserID: userID,
		}

		err = c.usecase.MoveTaskToAnotherList(ctx, taskInput)

		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToMoveTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task moved to another list", taskInput, slog.String(key.TaskID, taskInput.ID))
		}
	}
}

func (c *taskController) CompleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.MarkAsCompleted"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryTaskID)
			return
		}

		taskInput := model.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		err = c.usecase.CompleteTask(ctx, taskInput)

		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToCompleteTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task completed", nil, slog.String(key.TaskID, taskID))
		}
	}
}

func (c *taskController) ArchiveTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.MarkAsArchived"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
			return
		}

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyQueryTaskID)
			return
		}

		taskInput := model.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		err = c.usecase.ArchiveTask(ctx, taskInput)

		switch {
		case errors.Is(err, le.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, le.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToDeleteTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task deleted", taskID, slog.String(key.TaskID, taskID))
		}
	}
}
