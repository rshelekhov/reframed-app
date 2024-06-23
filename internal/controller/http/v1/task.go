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
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
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

		taskResponse, err := c.usecase.CreateTask(ctx, taskInput)
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

func (c *taskController) CreateTaskInDefaultList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.CreateTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		taskInput := &model.TaskRequestData{}
		if err = decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.UserID = userID

		taskResponse, err := c.usecase.CreateTask(ctx, taskInput)
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

func (c *taskController) GetTaskByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTaskByID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		taskID := chi.URLParam(r, key.TaskID)

		taskInput := model.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		taskResp, err := c.usecase.GetTaskByID(ctx, taskInput)
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

func (c *taskController) GetTasksByUserID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTasksByUserID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		pagination := ParseLimitAndAfterID(r)

		tasksResp, err := c.usecase.GetTasksByUserID(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks found", nil,
				slog.Int(key.Count, len(tasksResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "tasks found", tasksResp,
				slog.Int(key.Count, len(tasksResp)))
		}
	}
}

func (c *taskController) GetTasksByListID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTasksByListID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		listID := chi.URLParam(r, key.ListID)

		tasksInput := model.TaskRequestData{
			ListID: listID,
			UserID: userID,
		}

		tasksResp, err := c.usecase.GetTasksByListID(ctx, tasksInput)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks found for the list", nil,
				slog.Int(key.Count, len(tasksResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "tasks for the list found", tasksResp,
				slog.Int(key.Count, len(tasksResp)))
		}
	}
}

func (c *taskController) GetTasksGroupedByHeadings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTasksGroupedByHeadings"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		listID := chi.URLParam(r, key.ListID)

		tasksInput := model.TaskRequestData{
			ListID: listID,
			UserID: userID,
		}

		tasksResp, err := c.usecase.GetTasksGroupedByHeadings(ctx, tasksInput)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks grouped by headings found", nil,
				slog.Int(key.Count, len(tasksResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
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
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		tasksResp, err := c.usecase.GetTasksForToday(ctx, userID)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks found for today", nil,
				slog.Int(key.Count, len(tasksResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
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
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToParseQueryParams, err)
		}

		tasksResp, err := c.usecase.GetUpcomingTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no upcoming tasks found", nil,
				slog.Int(key.Count, len(tasksResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
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
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		pagination := ParseLimitAndAfterID(r)

		tasksResp, err := c.usecase.GetOverdueTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no overdue tasks found", nil,
				slog.Int(key.Count, len(tasksResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
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
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		pagination := ParseLimitAndAfterID(r)

		tasksResp, err := c.usecase.GetTasksForSomeday(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no tasks found for someday", nil,
				slog.Int(key.Count, len(tasksResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
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
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToParseQueryParams, err)
		}

		tasksResp, err := c.usecase.GetCompletedTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no completed tasks found", nil,
				slog.Int(key.Count, len(tasksResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "completed tasks found", tasksResp)
		}
	}
}

func (c *taskController) GetArchivedTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetArchivedTasks"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, le.ErrFailedToParseQueryParams, err)
		}

		tasksResp, err := c.usecase.GetArchivedTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, le.ErrNoTasksFound):
			handleResponseSuccess(w, r, log, "no archived tasks found", nil,
				slog.Int(key.Count, len(tasksResp)))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetData, err)
		default:
			handleResponseSuccess(w, r, log, "archived tasks found", tasksResp)
		}
	}
}

func (c *taskController) UpdateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.UpdateTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		taskID := chi.URLParam(r, key.TaskID)

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
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateTask, err)
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
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		taskID := chi.URLParam(r, key.TaskID)

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
		case errors.Is(err, le.ErrInvalidTaskTimeRange):
			handleResponseError(w, r, log, http.StatusBadRequest, le.ErrInvalidTaskTimeRange)
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToUpdateTask, err)
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
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
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

		taskResponse, err := c.usecase.MoveTaskToAnotherList(ctx, taskInput)
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

func (c *taskController) MoveTaskToAnotherHeading() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.MoveTaskToAnotherHeading"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
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

		taskResponse, err := c.usecase.MoveTaskToAnotherHeading(ctx, taskInput)
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

func (c *taskController) CompleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.CompleteTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		taskID := chi.URLParam(r, key.TaskID)

		taskInput := model.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		taskResponse, err := c.usecase.CompleteTask(ctx, taskInput)
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

func (c *taskController) ArchiveTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.ArchiveTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)

		userID, err := c.jwt.GetUserID(ctx)
		switch {
		case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
			handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()),
				slog.String(key.UserID, userID))
		case err != nil:
			handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		}

		taskID := chi.URLParam(r, key.TaskID)

		taskInput := model.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		taskResponse, err := c.usecase.ArchiveTask(ctx, taskInput)
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
