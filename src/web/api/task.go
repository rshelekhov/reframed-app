package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/logger"
	"github.com/rshelekhov/reframed/src/models"
	"github.com/rshelekhov/reframed/src/server/middleware/jwtoken/service"
	"github.com/rshelekhov/reframed/src/storage"
	"github.com/segmentio/ksuid"
	"log/slog"
	"net/http"
)

type TaskHandler struct {
	logger         logger.Interface
	tokenAuth      *service.JWTokenService
	taskStorage    storage.TaskStorage
	headingStorage storage.HeadingStorage
	tagStorage     storage.TagStorage
}

func NewTaskHandler(
	log logger.Interface,
	tokenAuth *service.JWTokenService,
	taskStorage storage.TaskStorage,
	headingStorage storage.HeadingStorage,
	tagStorage storage.TagStorage,
) *TaskHandler {
	return &TaskHandler{
		logger:         log,
		tokenAuth:      tokenAuth,
		taskStorage:    taskStorage,
		headingStorage: headingStorage,
		tagStorage:     tagStorage,
	}
}

func (h *TaskHandler) CreateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.CreateTask"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID := chi.URLParam(r, c.ListIDKey)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryListID)
			return
		}

		headingID := chi.URLParam(r, c.HeadingIDKey)
		if headingID == "" {
			headingID, err = h.headingStorage.GetDefaultHeadingID(r.Context(), listID, userID)
			if errors.Is(err, c.ErrHeadingNotFound) {
				handleResponseError(w, r, log, http.StatusNotFound, c.ErrHeadingNotFound)
				return
			}
			if err != nil {
				handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
				return
			}
		}

		task := &models.Task{}

		if err = DecodeAndValidateJSON(w, r, log, task); err != nil {
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
			HeadingID:   headingID,
			UserID:      userID,
			Tags:        task.Tags,
		}

		if err = h.taskStorage.CreateTask(r.Context(), newTask); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToCreateTask, err)
			return
		}

		for _, tag := range task.Tags {
			if err = h.tagStorage.CreateTagIfNotExists(r.Context(), tag, userID); err != nil {
				handleInternalServerError(w, r, log, c.ErrFailedToCreateTag, err)
				return
			}
		}

		if err = h.tagStorage.LinkTagsToTask(r.Context(), newTask.ID, task.Tags); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToLinkTagsToTask, err)
			return
		}

		handleResponseCreated(w, r, log, "task created", newTask, slog.String(c.TaskIDKey, newTask.ID))
	}
}

func (h *TaskHandler) GetTaskByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetTaskByID"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID := chi.URLParam(r, c.TaskIDKey)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryTaskID)
			return
		}

		task, err := h.taskStorage.GetTaskByID(r.Context(), taskID, userID)
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

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		pagination := ParseLimitAndAfterID(r)

		tasks, err := h.taskStorage.GetTasksByUserID(r.Context(), userID, pagination)
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
			)
		}
	}
}

func (h *TaskHandler) GetTasksByListID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetTasksByListID"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID := chi.URLParam(r, c.ListIDKey)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryListID)
			return
		}

		tasks, err := h.taskStorage.GetTasksByListID(r.Context(), listID, userID)
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
			)
		}
	}
}

func (h *TaskHandler) GetTasksGroupedByHeadings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetTasksGroupedByHeadings"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		listID := chi.URLParam(r, c.ListIDKey)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryListID)
			return
		}

		tasks, err := h.taskStorage.GetTasksGroupedByHeadings(r.Context(), listID, userID)
		switch {
		case errors.Is(err, c.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks grouped by headings found", tasks, slog.Int(c.CountKey, len(tasks)))
		}
	}
}

func (h *TaskHandler) GetTasksForToday() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetTasksForToday"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		tasks, err := h.taskStorage.GetTasksForToday(r.Context(), userID)
		switch {
		case errors.Is(err, c.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks for today found", tasks, slog.Int(c.CountKey, len(tasks)))
		}
	}
}

func (h *TaskHandler) GetUpcomingTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetUpcomingTasks"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToParseQueryParams, err)
			return
		}

		tasks, err := h.taskStorage.GetUpcomingTasks(r.Context(), userID, pagination)
		switch {
		case errors.Is(err, c.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "upcoming tasks found", tasks, slog.Int(c.CountKey, len(tasks)))
		}
	}
}

func (h *TaskHandler) GetOverdueTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetOverdueTasks"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		pagination := ParseLimitAndAfterID(r)

		tasks, err := h.taskStorage.GetOverdueTasks(r.Context(), userID, pagination)
		switch {
		case errors.Is(err, c.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "overdue tasks found", tasks, slog.Int(c.CountKey, len(tasks)))
		}
	}
}

func (h *TaskHandler) GetTasksForSomeday() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetTasksForSomeday"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		pagination := ParseLimitAndAfterID(r)

		tasks, err := h.taskStorage.GetTasksForSomeday(r.Context(), userID, pagination)
		switch {
		case errors.Is(err, c.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "tasks for someday found", tasks, slog.Int(c.CountKey, len(tasks)))
		}
	}
}

func (h *TaskHandler) GetCompletedTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetCompletedTasks"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToParseQueryParams, err)
			return
		}

		tasks, err := h.taskStorage.GetCompletedTasks(r.Context(), userID, pagination)
		switch {
		case errors.Is(err, c.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "completed tasks found", tasks, slog.Int(c.CountKey, len(tasks)))
		}
	}
}

func (h *TaskHandler) GetArchivedTasks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.GetArchivedTasks"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		pagination, err := ParseLimitAndAfterDate(r)
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToParseQueryParams, err)
			return
		}

		tasks, err := h.taskStorage.GetArchivedTasks(r.Context(), userID, pagination)
		switch {
		case errors.Is(err, c.ErrNoTasksFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrNoTasksFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToGetData, err)
			return
		default:
			handleResponseSuccess(w, r, log, "archived tasks found", tasks, slog.Int(c.CountKey, len(tasks)))
		}
	}
}

func (h *TaskHandler) UpdateTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.UpdateTask"

		log := logger.LogWithRequest(h.logger, op, r)
		task := &models.UpdateTask{}

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID := chi.URLParam(r, c.TaskIDKey)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryTaskID)
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
			ListID:      task.ListID,
			HeadingID:   task.HeadingID,
			UserID:      userID,
			Tags:        task.Tags,
		}

		currentTags, err := h.tagStorage.GetTagsByTaskID(r.Context(), taskID)
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
			if err = h.tagStorage.CreateTagIfNotExists(r.Context(), tag, userID); err != nil {
				handleInternalServerError(w, r, log, c.ErrFailedToCreateTag, err)
				return
			}
		}

		if err = h.tagStorage.LinkTagsToTask(r.Context(), updatedTask.ID, tagsToAdd); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToLinkTagsToTask, err)
			return
		}

		if err = h.tagStorage.UnlinkTagsFromTask(r.Context(), updatedTask.ID, tagsToRemove); err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToDeleteTag, err)
			return
		}

		err = h.taskStorage.UpdateTask(r.Context(), updatedTask)
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

		log := logger.LogWithRequest(h.logger, op, r)
		task := &models.UpdateTask{}

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID := chi.URLParam(r, c.TaskIDKey)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryTaskID)
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

		err = h.taskStorage.UpdateTask(r.Context(), updatedTask)
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

func (h *TaskHandler) MoveTaskToAnotherList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.MoveTaskToAnotherList"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID := chi.URLParam(r, c.TaskIDKey)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryTaskID)
			return
		}

		listID := r.URL.Query().Get(c.ListIDKey)
		if listID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryListID)
			return
		}

		err = h.taskStorage.MoveTaskToAnotherList(r.Context(), listID, taskID, userID)
		switch {
		case errors.Is(err, c.ErrTaskNotFound):
			handleResponseError(w, r, log, http.StatusNotFound, c.ErrTaskNotFound)
			return
		case err != nil:
			handleInternalServerError(w, r, log, c.ErrFailedToMoveTask, err)
			return
		default:
			handleResponseSuccess(w, r, log, "task completed", nil, slog.String(c.TaskIDKey, taskID))
		}
	}
}

func (h *TaskHandler) CompleteTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.CompleteTask"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID := chi.URLParam(r, c.TaskIDKey)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryTaskID)
			return
		}

		err = h.taskStorage.CompleteTask(r.Context(), taskID, userID)
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

func (h *TaskHandler) ArchiveTask() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "task.handlers.ArchiveTask"

		log := logger.LogWithRequest(h.logger, op, r)

		_, claims, err := service.GetTokenFromContext(r.Context())
		if err != nil {
			handleInternalServerError(w, r, log, c.ErrFailedToGetAccessToken, err)
			return
		}
		userID := claims[c.ContextUserID].(string)

		taskID := chi.URLParam(r, c.TaskIDKey)
		if taskID == "" {
			handleResponseError(w, r, log, http.StatusBadRequest, c.ErrEmptyQueryTaskID)
			return
		}

		err = h.taskStorage.ArchiveTask(r.Context(), taskID, userID)
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
