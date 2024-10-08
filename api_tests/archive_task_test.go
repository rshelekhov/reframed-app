package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/segmentio/ksuid"
	"net/http"
	"net/url"
	"testing"
)

func TestArchiveTask_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	user := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := user.Value(jwtoken.AccessTokenKey).String().Raw()

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task
	task := e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()

	// Archive task
	archivedTask := e.PATCH("/user/tasks/{task_id}/archive", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	taskStatusID := archivedTask.Value(key.Data).Object().Value(key.StatusID).Raw()

	// Get status
	taskStatus := e.GET("/statuses/{status_id}", taskStatusID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	taskStatusTitle := taskStatus.Value(key.Data).Object().Value(key.Title).String().Raw()

	if taskStatusTitle != model.StatusArchived.String() {
		t.Errorf("expected task status to be %s, but got %s", model.StatusArchived.String(), taskStatusTitle)
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestArchiveTask_NotFound(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	user := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := user.Value(jwtoken.AccessTokenKey).String().Raw()

	taskID := ksuid.New().String()

	e.PATCH("/user/tasks/{task_id}/archive", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusNotFound)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
