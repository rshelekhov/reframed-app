package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"net/http"
	"net/url"
	"testing"
)

func TestCreateTaskInDefaultList_HappyPath(t *testing.T) {
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

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task
	e.POST("/user/lists/default").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().NotEmpty()

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestCreateTaskOnSpecificList_HappyPath(t *testing.T) {
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

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task
	task := e.POST("/user/lists/{list_id}/tasks/", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()

	getTask := e.GET("/user/tasks/{task_id}/", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	taskList := getTask.Value(key.Data).Object().Value(key.ListID).String().Raw()

	if taskList != listID {
		t.Errorf("expected task list to be %s, but got %s", listID, taskList)
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestCreateTaskOnSpecificHeading_HappyPath(t *testing.T) {
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

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task
	task := e.POST("/user/lists/{list_id}/headings/{heading_id}/", listID, headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	taskHeadingID := task.Value(key.Data).Object().Value(key.HeadingID).String().Raw()

	if taskHeadingID != headingID {
		t.Errorf("expected task heading to be %s, but got %s", headingID, taskHeadingID)
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestCreateTask_InvalidListID(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register first user
	user1 := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken1 := user1.Value(jwtoken.AccessTokenKey).String().Raw()

	// Register second user
	user2 := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken2 := user2.Value(jwtoken.AccessTokenKey).String().Raw()

	// Create list
	list2 := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken2).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID2 := list2.Value(key.Data).Object().Value(key.ListID).String().Raw()

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task with invalid listID
	e.POST("/user/lists/{list_id}/tasks/", listID2).
		WithHeader("Authorization", "Bearer "+accessToken1).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object()

	// Cleanup the SSO gRPC service storage after testing
	responses := []*httpexpect.Object{user1, user2}
	for _, resp := range responses {
		cleanupAuthService(e, resp)
	}
}

func TestCreateTask_InvalidHeadingID(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register first user
	user1 := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken1 := user1.Value(jwtoken.AccessTokenKey).String().Raw()

	// Create list for the first user
	list := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken1).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID := list.Value(key.Data).Object().Value(key.ListID).String().Raw()

	// Create heading for the first list
	heading := e.POST("/user/lists/{list_id}/headings/", listID).
		WithHeader("Authorization", "Bearer "+accessToken1).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	headingID := heading.Value(key.Data).Object().Value(key.HeadingID).String().Raw()

	// Register second user
	user2 := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken2 := user2.Value(jwtoken.AccessTokenKey).String().Raw()

	// Create list
	list2 := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken2).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID2 := list2.Value(key.Data).Object().Value(key.ListID).String().Raw()

	fakeTask := randomFakeTask(upcomingTasks, "", "")

	// Create task with invalid headingID
	e.POST("/user/lists/{list_id}/headings/{heading_id}/", listID2, headingID).
		WithHeader("Authorization", "Bearer "+accessToken2).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object()

	// Cleanup the SSO gRPC service storage after testing
	responses := []*httpexpect.Object{user1, user2}
	for _, resp := range responses {
		cleanupAuthService(e, resp)
	}
}
