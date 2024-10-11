package api_tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/jwtauth"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/stretchr/testify/require"
)

func TestGetTasksForToday_HappyPath(t *testing.T) {
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
	_ = createTasks(e, accessToken, todayTasks, lists, numberOfTasks)

	// Get tasks for today
	tasks := e.GET("/user/tasks/today").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	totalTasks := countTasksInGroups(t, tasks, false)
	require.Equal(t, numberOfLists*numberOfTasks, totalTasks)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestGetTasksForToday_NotFound(t *testing.T) {
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

	// Get tasks for today
	tasks := e.GET("/user/tasks/today").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	printDataToJSON(t, tasks)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
