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

func TestCreateTaskInDefaultList_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
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

	fakeTask := randomFakeTask(true, true, true, true, true, "", "")

	// Create task
	e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()
}

func TestCreateTaskInDefaultList_ListNotFound(t *testing.T) {
	u := url.URL{
		Scheme: "http",
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

	// Get default list
	l := e.GET("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	defaultListID := l.Value("data").Object().Value("list_id").String().Raw()

	// Remove default list
	e.DELETE("/user/lists/{list_id}").
		WithPath("list_id", defaultListID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK)

	fakeTask := randomFakeTask(false, false, false, false, false, "", "")

	// Create task
	e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusNotFound)
}

func TestCreateTaskOnSpecificList_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
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

	l := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID := l.Value("data").Object().Value("list_id").String().Raw()

	fakeTask := randomFakeTask(true, true, true, true, true, "", "")

	// Create task
	task := e.POST("/user/lists/{list_id}/tasks/").
		WithPath("list_id", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	task.NotEmpty()

	taskID := task.Value("data").Object().Value("task_id").String().Raw()

	getTask := e.GET("/user/tasks/{task_id}/").
		WithPath("task_id", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	getTask.NotEmpty()

	taskList := getTask.Value("data").Object().Value("list_id").String().Raw()

	require.Equal(t, listID, taskList)
}
