package api_tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/jwtauth"
	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/segmentio/ksuid"
)

func TestUpdateTask_HappyPath(t *testing.T) {
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

	accessToken := user.Value(jwtauth.AccessTokenKey).String().Raw()

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task
	task := e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()

	// Update task
	e.PATCH("/user/tasks/{task_id}", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.TaskRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestUpdateTask_NotFound(t *testing.T) {
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

	accessToken := user.Value(jwtauth.AccessTokenKey).String().Raw()

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task
	e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().NotEmpty()

	fakeID := ksuid.New().String()

	// Update task
	e.PATCH("/user/tasks/{task_id}", fakeID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.TaskRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusNotFound)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
