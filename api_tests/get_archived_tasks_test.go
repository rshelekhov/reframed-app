package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/stretchr/testify/require"
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
	user := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := user.Value(jwtoken.AccessTokenKey).String().Raw()

	numberOfLists := 3
	numberOfTasks := 3

	// Create three lists
	lists := createLists(e, accessToken, numberOfLists)

	// Create three tasks in each list
	tasks := createTasks(e, accessToken, overdueTasks, lists, numberOfTasks)

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
	archivedTasks := e.GET("/user/tasks/archived").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	// Check that returned the same amount of tasks as was archived
	totalArchivedTasks := countTasksInGroups(t, archivedTasks, false)

	require.Equal(t, numberOfLists*numberOfTasks, totalArchivedTasks)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetArchivedTasks_NotFound(t *testing.T) {
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

	// Get archived tasks
	archivedTasks := e.GET("/user/tasks/archived").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	printDataToJSON(t, archivedTasks)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
