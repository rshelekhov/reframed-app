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

func TestMoveTaskToAnotherList_HappyPath(t *testing.T) {
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

	// Move task
	movedTask := e.PATCH("/user/tasks/{task_id}/move/list", taskID).
		WithQuery(key.ListID, listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	movedTaskListID := movedTask.Value(key.Data).Object().Value(key.ListID).String().Raw()

	if movedTaskListID != listID {
		t.Errorf("expected moved task list ID to be %s, but got %s", listID, movedTaskListID)
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestMoveTaskToAnotherList_FailCases(t *testing.T) {
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

	testCases := []struct {
		name   string
		taskID string
		listID string
		status int
	}{
		{
			name:   "Move task to another list when listID is empty",
			taskID: taskID,
			listID: "",
			status: http.StatusBadRequest,
		},
		{
			name:   "Move task to another list when task is not found",
			taskID: ksuid.New().String(),
			listID: listID,
			status: http.StatusNotFound,
		},
		{
			name:   "Move task to another list when list not found",
			taskID: taskID,
			listID: ksuid.New().String(),
			status: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e.PATCH("/user/tasks/{task_id}/move/list", tc.taskID).
				WithQuery(key.ListID, tc.listID).
				WithHeader("Authorization", "Bearer "+accessToken).
				Expect().
				Status(tc.status)
		})
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestMoveTaskToAnotherHeading_HappyPath(t *testing.T) {
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

	// Create heading
	heading := e.POST("/user/lists/{list_id}/headings/", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	headingID := heading.Value(key.Data).Object().Value(key.HeadingID).String().Raw()

	// Move task
	movedTask := e.PATCH("/user/tasks/{task_id}/move/heading", taskID).
		WithQuery(key.HeadingID, headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	movedTaskHeadingID := movedTask.Value(key.Data).Object().Value(key.HeadingID).String().Raw()

	if movedTaskHeadingID != headingID {
		t.Errorf("expected moved task heading ID to be %s, but got %s", headingID, movedTaskHeadingID)
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestMoveTaskToAnotherHeading_FailCases(t *testing.T) {
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

	// Create heading
	heading := e.POST("/user/lists/{list_id}/headings/", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	headingID := heading.Value(key.Data).Object().Value(key.HeadingID).String().Raw()

	testCases := []struct {
		name      string
		taskID    string
		listID    string
		headingID string
		status    int
	}{
		{
			name:      "Move task to another heading when headingID is empty",
			taskID:    taskID,
			listID:    listID,
			headingID: "",
			status:    http.StatusBadRequest,
		},
		{
			name:      "Move task to another heading when task is not found",
			taskID:    ksuid.New().String(),
			listID:    listID,
			headingID: headingID,
			status:    http.StatusNotFound,
		},
		{
			name:      "Move task to another heading when heading is not found",
			taskID:    taskID,
			listID:    listID,
			headingID: ksuid.New().String(),
			status:    http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e.PATCH("/user/tasks/{task_id}/move/heading", tc.taskID).
				WithQuery(key.HeadingID, tc.headingID).
				WithHeader("Authorization", "Bearer "+accessToken).
				Expect().
				Status(tc.status)
		})
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
