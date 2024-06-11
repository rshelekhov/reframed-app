package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func TestArchiveTask_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	r := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := r.Value(jwtoken.AccessTokenKey).String().Raw()

	fakeTask := randomFakeTask(true, true, true, true, true, "", "")

	// Create task
	task := e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	taskID := task.Value("data").Object().Value("task_id").String().Raw()

	// Archive task
	archivedTask := e.PATCH("/user/tasks/{task_id}/archive", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()

	taskStatusID := archivedTask.Value("data").Object().Value("status_id").Raw()

	// Get status
	taskStatus := e.GET("/statuses/{status_id}", taskStatusID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	taskStatusTitle := taskStatus.Value("data").Object().Value("title").String().Raw()

	require.Equal(t, taskStatusTitle, model.StatusArchived.String())
}

func TestArchiveTask_FailCases(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	r := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := r.Value(jwtoken.AccessTokenKey).String().Raw()

	testCases := []struct {
		name   string
		taskID string
		status int
	}{
		{
			name:   "Archive task with empty task_id",
			taskID: "",
			status: http.StatusBadRequest,
		},
		{
			name:   "Archive task when task not found",
			taskID: ksuid.New().String(),
			status: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e.PATCH("/user/tasks/{task_id}/archive", tc.taskID).
				WithHeader("Authorization", "Bearer "+accessToken).
				Expect().
				Status(tc.status)
		})
	}
}
