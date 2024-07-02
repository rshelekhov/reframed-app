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

func TestCompleteTask_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
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

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task
	task := e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()

	// Complete task
	completedTask := e.PATCH("/user/tasks/{task_id}/complete", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	taskStatusID := completedTask.Value(key.Data).Object().Value(key.StatusID).Raw()

	// Get status
	taskStatus := e.GET("/statuses/{status_id}", taskStatusID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	taskStatusTitle := taskStatus.Value(key.Data).Object().Value(key.Title).String().Raw()

	if taskStatusTitle != model.StatusCompleted.String() {
		t.Errorf("expected task status to be %s, but got %s", model.StatusCompleted.String(), taskStatusTitle)
	}
}

func TestCompleteTask_NotFound(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
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

	taskID := ksuid.New().String()

	e.PATCH("/user/tasks/{task_id}/complete", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusNotFound)

}
