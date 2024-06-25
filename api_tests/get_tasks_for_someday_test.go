package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func TestGetTasksForSomeday_HappyPath(t *testing.T) {
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
	_ = createTasks(e, accessToken, somedayTasks, lists, numberOfTasks)

	// Get tasks for someday
	tasks := e.GET("/user/tasks/someday").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	totalTasks := countTasksInGroups(t, tasks, false)

	require.Equal(t, numberOfLists*numberOfTasks, totalTasks)
}

func TestGetTasksForSomeday_WithPagination(t *testing.T) {
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
	_ = createTasks(e, accessToken, somedayTasks, lists, numberOfTasks)

	testCases := []struct {
		name          string
		limit         int
		expectedLists int
	}{
		{
			name:          "Get tasks for someday with limit = 2",
			limit:         2,
			expectedLists: 2,
		},
		{
			name:          "Get tasks for someday with limit = 0",
			limit:         0,
			expectedLists: numberOfLists,
		},
		{
			name:  "Get tasks for someday with limit = -1",
			limit: -1,
			// Limit = -1 means no limit, will be used the default value
			expectedLists: numberOfLists,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get tasks for someday
			tasks := e.GET("/user/tasks/someday").
				WithHeader("Authorization", "Bearer "+accessToken).
				WithQuery("limit", tc.limit).
				Expect().
				Status(http.StatusOK).
				JSON().Object()

			totalLists := countGroups(t, tasks, false)

			require.Equal(t, tc.expectedLists, totalLists)
		})
	}
}

func TestGetTasksForSomeday_NotFound(t *testing.T) {
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
	tasks := e.GET("/user/tasks/someday").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	printDataToJSON(t, tasks)
}
