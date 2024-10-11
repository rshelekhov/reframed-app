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
	"github.com/stretchr/testify/require"
)

func TestGetCompletedTasks_HappyPath(t *testing.T) {
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

	numberOfLists := 3
	numberOfTasks := 3

	// Create three lists
	lists := createLists(e, accessToken, numberOfLists)

	// Create three tasks in each list
	tasks := createTasks(e, accessToken, overdueTasks, lists, numberOfTasks)

	// Complete tasks
	for _, task := range tasks {
		taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()
		e.PATCH("/user/tasks/{task_id}/complete", taskID).
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object().NotEmpty()
	}

	// Get completed tasks
	completedTasks := e.GET("/user/tasks/completed").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	// Check that returned the same amount of tasks as was completed
	totalCompletedTasks := countTasksInGroups(t, completedTasks, false)

	require.Equal(t, numberOfTasks*numberOfLists, totalCompletedTasks)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetCompletedTasks_WithLimit(t *testing.T) {
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

	numberOfLists := 3
	numberOfTasks := 3

	// Create three lists
	lists := createLists(e, accessToken, numberOfLists)

	// Create three tasks in each list
	tasks := createTasks(e, accessToken, overdueTasks, lists, numberOfTasks)

	// Complete tasks
	for _, task := range tasks {
		taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()
		e.PATCH("/user/tasks/{task_id}/complete", taskID).
			WithHeader("Authorization", "Bearer "+accessToken).
			Expect().
			Status(http.StatusOK).
			JSON().Object().NotEmpty()
	}

	testCases := []struct {
		name           string
		limit          int
		expectedGroups int
	}{
		{
			name:           "Get completed tasks with limit = 2",
			limit:          1,
			expectedGroups: 1,
		},
		{
			name:           "Get completed tasks with limit = 0",
			limit:          0,
			expectedGroups: 1,
		},
		{
			name:  "Get completed tasks with limit = -1",
			limit: -1,
			// Limit = -1 means no limit, will be used the default value
			expectedGroups: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get completed tasks
			completedTasks := e.GET("/user/tasks/completed").
				WithHeader("Authorization", "Bearer "+accessToken).
				WithQuery(key.Limit, tc.limit).
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			totalGroups := countGroups(t, completedTasks, false)
			require.Equal(t, tc.expectedGroups, totalGroups)
		})
	}
}

func TestGetCompletedTasks_NotFound(t *testing.T) {
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

	// Get completed tasks
	completedTasks := e.GET("/user/tasks/completed").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	printDataToJSON(t, completedTasks)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
