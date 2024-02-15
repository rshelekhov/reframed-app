package v1

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/rshelekhov/reframed/internal/domain"
	"github.com/rshelekhov/reframed/pkg/constants/key"
	"github.com/rshelekhov/reframed/pkg/httpserver/middleware/jwtoken"
	"github.com/rshelekhov/reframed/pkg/logger"
	"log/slog"
	"net/http"
)

type taskController struct {
	logger  logger.Interface
	jwt     *jwtoken.TokenService
	usecase domain.TaskUsecase
}

func NewTaskRoutes(
	r *chi.Mux,
	log logger.Interface,
	jwt *jwtoken.TokenService,
	usecase domain.TaskUsecase,
) {
	c := &taskController{
		logger:  log,
		jwt:     jwt,
		usecase: usecase,
	}

	r.Group(func(r chi.Router) {
		r.Use(jwtoken.Verifier(jwt))
		r.Use(jwtoken.Authenticator())

		r.Route("/user/lists/{list_id}", func(r chi.Router) {
			r.Get("/tasks", c.GetTasksByListID())
			r.Post("/tasks", c.CreateTask())

			r.Route("/headings", func(r chi.Router) {
				r.Get("/tasks", c.GetTasksGroupedByHeadings())
				r.Post("/{heading_id}", c.CreateTask())
			})
		})

		r.Route("/user/tasks", func(r chi.Router) {
			r.Get("/", c.GetTasksByUserID())
			r.Get("/today", c.GetTasksForToday())      // grouped by list title
			r.Get("/upcoming", c.GetUpcomingTasks())   // grouped by start_date
			r.Get("/overdue", c.GetOverdueTasks())     // grouped by list title
			r.Get("/someday", c.GetTasksForSomeday())  // tasks without start_date, grouped by list title
			r.Get("/completed", c.GetCompletedTasks()) // grouped by month
			r.Get("/archived", c.GetArchivedTasks())   // grouped by month

			r.Route("/{task_id}", func(r chi.Router) {
				r.Get("/", c.GetTaskByID())
				r.Put("/", c.UpdateTask())
				r.Put("/time", c.UpdateTaskTime())
				r.Put("/move", c.MoveTaskToAnotherList())
				r.Put("/complete", c.CompleteTask())
				r.Delete("/", c.ArchiveTask())
			})
		})
	})
}

func (c *taskController) CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.CreateTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		headingID := chi.URLParam(r, key.HeadingID)

		taskInput := &domain.TaskRequestData{}
		if err := decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.HeadingID = headingID
		taskInput.ListID = listID
		taskInput.UserID = userID

		taskID, err := c.usecase.CreateTask(ctx, taskInput)
		switch {
		case errors.Is(err, domain.ErrHeadingNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrHeadingNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToCreateTask, err)
			return
		default:
			handleResponseCreated(w, r, log, "task created", domain.TaskResponseData{ID: taskID}, slog.String(key.TaskID, taskID))
		}
	}
}

func (c *taskController) GetTaskByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.GetTaskByID"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryTaskID)
			return
		}

		taskInput := domain.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		taskResp, err := c.usecase.GetTaskByID(ctx, taskInput)
		switch {
		case errors.Is(err, domain.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		pagination := ParseLimitAndAfterID(r)

		tasksResp, err := c.usecase.GetTasksByUserID(ctx, userID, pagination)
		switch {
		case errors.Is(err, domain.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		tasksInput := domain.TaskRequestData{
			ListID: listID,
			UserID: userID,
		}

		tasksResp, err := c.usecase.GetTasksByListID(ctx, tasksInput)
		switch {
		case errors.Is(err, domain.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		listID := chi.URLParam(r, key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		tasksInput := domain.TaskRequestData{
			ListID: listID,
			UserID: userID,
		}

		tasksResp, err := c.usecase.GetTasksGroupedByHeadings(ctx, tasksInput)
		switch {
		case errors.Is(err, domain.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		tasksResp, err := c.usecase.GetTasksForToday(ctx, userID)
		switch {
		case errors.Is(err, domain.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToParseQueryParams, err)
			return
		}

		tasksResp, err := c.usecase.GetUpcomingTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, domain.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		pagination := ParseLimitAndAfterID(r)

		tasksResp, err := c.usecase.GetOverdueTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, domain.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		pagination := ParseLimitAndAfterID(r)

		tasksResp, err := c.usecase.GetTasksForSomeday(ctx, userID, pagination)
		switch {
		case errors.Is(err, domain.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToParseQueryParams, err)
			return
		}

		tasksResp, err := c.usecase.GetCompletedTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, domain.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, domain.ErrFailedToParseQueryParams, err)
			return
		}

		tasksResp, err := c.usecase.GetArchivedTasks(ctx, userID, pagination)
		switch {
		case errors.Is(err, domain.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToGetData, err)
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
		userID := jwtoken.GetUserID(ctx).(string)

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryTaskID)
			return
		}

		taskInput := &domain.TaskRequestData{}
		if err := decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.ID = taskID
		taskInput.UserID = userID

		err := c.usecase.UpdateTask(ctx, taskInput)
		switch {
		case errors.Is(err, domain.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToUpdateTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task updated", taskID, slog.String(key.TaskID, taskID))
		}

	}
}

func (c *taskController) UpdateTaskTime() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.UpdateTaskTimes"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryTaskID)
			return
		}

		taskInput := &domain.TaskRequestData{}
		if err := decodeAndValidateJSON(w, r, log, taskInput); err != nil {
			return
		}

		taskInput.ID = taskID
		taskInput.UserID = userID

		err := c.usecase.UpdateTaskTime(ctx, taskInput)
		switch {
		case errors.Is(err, domain.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToUpdateTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task updated", taskID, slog.String(key.TaskID, taskID))
		}
	}
}

func (c *taskController) MoveTaskToAnotherList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.MoveTaskToAnotherList"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryTaskID)
			return
		}

		listID := r.URL.Query().Get(key.ListID)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryListID)
			return
		}

		taskInput := domain.TaskRequestData{
			ID:     taskID,
			ListID: listID,
			UserID: userID,
		}

		err := c.usecase.MoveTaskToAnotherList(ctx, taskInput)
		switch {
		case errors.Is(err, domain.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToMoveTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task moved to another list", taskID, slog.String(key.TaskID, taskID))
		}
	}
}

func (c *taskController) CompleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.CompleteTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryTaskID)
			return
		}

		taskInput := domain.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		err := c.usecase.CompleteTask(ctx, taskInput)
		switch {
		case errors.Is(err, domain.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToCompleteTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task completed", nil, slog.String(key.TaskID, taskID))
		}
	}
}

func (c *taskController) ArchiveTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.controller.ArchiveTask"

		ctx := r.Context()
		log := logger.LogWithRequest(c.logger, op, r)
		userID := jwtoken.GetUserID(ctx).(string)

		taskID := chi.URLParam(r, key.TaskID)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, domain.ErrEmptyQueryTaskID)
			return
		}

		taskInput := domain.TaskRequestData{
			ID:     taskID,
			UserID: userID,
		}

		err := c.usecase.ArchiveTask(ctx, taskInput)
		switch {
		case errors.Is(err, domain.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, domain.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, domain.ErrFailedToDeleteTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task deleted", domain.Task{ID: taskID}, slog.String(key.TaskID, taskID))
		}
	}
}
