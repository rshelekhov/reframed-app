package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"net/http"
	"net/url"
	"testing"
)

func TestGetArchivedTasks_HappyPath(t *testing.T) {
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

	// Create three lists
	lists := createLists(e, accessToken, 3)

	// Create three tasks in each list
	tasks := createTasks(e, accessToken, lists, 3)

	// Archive tasks
	for _, task := range tasks {
		taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()
		e.PATCH("/user/tasks/{task_id}/archive", taskID).
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object().NotEmpty()
	}

	// Get archived tasks
	e.GET("/user/tasks/archived").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON()

	// TODO: посчитать, что количество задач соответствует тому, что было заархивировано
}
