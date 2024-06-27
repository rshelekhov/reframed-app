package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func TestGetOverdueTasks_HappyPath(t *testing.T) {
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

	numberOfLists := 3
	numberOfTasks := 3

	// Create three lists
	lists := createLists(e, accessToken, numberOfLists)

	// Create three tasks in each list
	_ = createTasks(e, accessToken, overdueTasks, lists, numberOfTasks)

	// Get tasks for someday
	tasks := e.GET("/user/tasks/overdue").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	printDataToJSON(t, tasks)

	totalTasks := countTasksInGroups(t, tasks, false)
	require.Equal(t, numberOfLists*numberOfTasks, totalTasks)
}

func TestGetOverdueTasks_WithLimit(t *testing.T) {
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

	numberOfLists := 3
	numberOfTasks := 3

	// Create three lists
	lists := createLists(e, accessToken, numberOfLists)

	// Create three tasks in each list
	_ = createTasks(e, accessToken, overdueTasks, lists, numberOfTasks)

	testCases := []struct {
		name          string
		limit         int
		expectedLists int
	}{
		{
			name:          "Get overdue tasks with limit = 2",
			limit:         2,
			expectedLists: 2,
		},
		{
			name:          "Get overdue tasks with limit = 0",
			limit:         0,
			expectedLists: numberOfLists,
		},
		{
			name:  "Get overdue tasks with limit = -1",
			limit: -1,
			// Limit = -1 means no limit, will be used the default value
			expectedLists: numberOfLists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get tasks for someday
			tasks := e.GET("/user/tasks/overdue").
				WithHeader("Authorization", "Bearer "+accessToken).
				WithQuery(key.Limit, tc.limit).
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			totalLists := countGroups(t, tasks, false)
			require.Equal(t, tc.expectedLists, totalLists)
		})
	}
}

func TestGetOverdueTasks_WithPagination(t *testing.T) {
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

	numberOfLists := 3
	numberOfTasks := 3

	// Create three lists
	lists := createLists(e, accessToken, numberOfLists)

	// Create three tasks in each list
	_ = createTasks(e, accessToken, overdueTasks, lists, numberOfTasks)

	limit := 1

	// First request to get tasks for someday
	response := e.GET("/user/tasks/overdue").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithQuery(key.Limit, limit).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	// Ensure we received tasks
	totalGroups := countGroups(t, response, false)
	require.Equal(t, limit, totalGroups)

	// Extract the last list_id from the response
	lastGroup := response.Value(key.Data).Array().Last().Object()
	lastListID := lastGroup.Value(key.ListID).String().Raw()

	// Second request to get tasks for someday with afterID set to last list_id from the first response
	nextResponse := e.GET("/user/tasks/overdue").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithQuery(key.Limit, limit).
		WithQuery(key.Cursor, lastListID).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	// Count the number of task groups (lists) in the second response
	nextTotalGroups := countGroups(t, nextResponse, false)
	expectedNextGroups := numberOfLists - limit - 1 // Skip limit and the last group of the first response
	require.Equal(t, expectedNextGroups, nextTotalGroups)
}

func TestGetOverdueTasks_NotFound(t *testing.T) {
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

	// Get upcoming tasks
	tasks := e.GET("/user/tasks/overdue").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	printDataToJSON(t, tasks)
}
