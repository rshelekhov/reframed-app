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

func TestGetTasksByUserID_HappyPath(t *testing.T) {
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
	_ = createTasks(e, accessToken, upcomingTasks, lists, numberOfTasks)

	// Get tasks
	e.GET("/user/tasks").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetTasksByUserID_WithLimit(t *testing.T) {
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
	_ = createTasks(e, accessToken, upcomingTasks, lists, numberOfTasks)

	testCases := []struct {
		name          string
		limit         int
		expectedTasks int
	}{
		{
			name:          "Get tasks for someday with limit = 2",
			limit:         2,
			expectedTasks: 2,
		},
		{
			name:          "Get tasks for someday with limit = 0",
			limit:         0,
			expectedTasks: numberOfLists * numberOfTasks,
		},
		{
			name:  "Get tasks for someday with limit = -1",
			limit: -1,
			// Limit = -1 means no limit, will be used the default value
			expectedTasks: numberOfLists * numberOfTasks,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get tasks for someday
			tasks := e.GET("/user/tasks/").
				WithHeader("Authorization", "Bearer "+accessToken).
				WithQuery(key.Limit, tc.limit).
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			totalTasks := countTasks(t, tasks, false)
			require.Equal(t, tc.expectedTasks, totalTasks)
		})
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetTasksByUserID_WithPagination(t *testing.T) {
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
	_ = createTasks(e, accessToken, upcomingTasks, lists, numberOfTasks)

	limit := 1

	// First request to get tasks by userID
	firstResponse := e.GET("/user/tasks").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithQuery(key.Limit, limit).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	// Ensure we received tasks
	totalTasks := countTasks(t, firstResponse, false)
	require.Equal(t, limit, totalTasks)

	// Extract the last task_id from the firstResponse
	lastTask := firstResponse.Value(key.Data).Array().Last().Object()
	lastTaskID := lastTask.Value(key.TaskID).String().Raw()

	// Second request to get tasks for someday with cursor set to last task_id from the first firstResponse
	secondResponse := e.GET("/user/tasks").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithQuery(key.Cursor, lastTaskID).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	// Count the number of tasks in the second firstResponse
	nextTotalTasks := countTasks(t, secondResponse, false)
	expectedNextTasks := numberOfLists*numberOfTasks - 1 // Skip the last task of the first firstResponse
	require.Equal(t, expectedNextTasks, nextTotalTasks)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetTasksByUserID_NotFound(t *testing.T) {
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

	// Try to get tasks
	e.GET("/user/tasks").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetTasksByListID_HappyPath(t *testing.T) {
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

	list := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID := list.Value(key.Data).Object().Value(key.ListID).String().Raw()

	numberOfTasks := 3

	for i := 0; i < numberOfTasks; i++ {
		fakeTask := randomFakeTask(upcomingTasks, listID, "")

		e.POST("/user/lists/{list_id}/tasks", listID).
			WithHeader("Authorization", "Bearer "+accessToken).
			WithJSON(fakeTask).
			Expect().
			Status(http.StatusCreated).
			JSON().Object().NotEmpty()
	}

	// Get tasks by listID
	tasks := e.GET("/user/lists/{list_id}/tasks", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	totalTasksInList := countTasks(t, tasks, false)

	require.Equal(t, numberOfTasks, totalTasksInList)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetTasksByListID_NotFound(t *testing.T) {
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

	// Create list
	list := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID := list.Value(key.Data).Object().Value(key.ListID).String().Raw()

	// Get tasks by listID
	e.GET("/user/lists/{list_id}/tasks", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetTasksGroupedByHeading_HappyPath(t *testing.T) {
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

	// Create list
	list := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID := list.Value(key.Data).Object().Value(key.ListID).String().Raw()

	numberOfTasks := 3

	for i := 0; i < numberOfTasks; i++ {
		fakeTask := randomFakeTask(upcomingTasks, listID, "")

		e.POST("/user/lists/{list_id}/tasks", listID).
			WithHeader("Authorization", "Bearer "+accessToken).
			WithJSON(fakeTask).
			Expect().
			Status(http.StatusCreated).
			JSON().Object().NotEmpty()
	}

	// Get tasks by listID
	tasks := e.GET("/user/lists/{list_id}/headings/tasks", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	totalTasks := countTasksInGroups(t, tasks, false)

	require.Equal(t, numberOfTasks, totalTasks)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetTasksGroupedByHeading_NotFound(t *testing.T) {
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

	// Create list
	list := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID := list.Value(key.Data).Object().Value(key.ListID).String().Raw()

	// Get tasks by listID
	tasks := e.GET("/user/lists/{list_id}/headings/tasks", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	printDataToJSON(t, tasks)

	totalTasks := countTasksInGroups(t, tasks, false)

	require.Equal(t, 0, totalTasks)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
