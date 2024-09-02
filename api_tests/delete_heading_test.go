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

func TestDeleteHeading_HappyPath(t *testing.T) {
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

	// Delete heading
	e.DELETE("/user/lists/{list_id}/headings/{heading_id}", listID, headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestDeleteHeading_DoubleDelete(t *testing.T) {
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

	// Create task on the heading
	fakeTask := randomFakeTask(upcomingTasks, "", "")

	task := e.POST("/user/lists/{list_id}/headings/{heading_id}/", listID, headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(fakeTask).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	// Delete heading
	e.DELETE("/user/lists/{list_id}/headings/{heading_id}", listID, headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK)

	// Try to delete heading again
	e.DELETE("/user/lists/{list_id}/headings/{heading_id}", listID, headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusNotFound)

	// Check that task was deleted
	taskID := task.Value(key.Data).Object().Value(key.TaskID).String().Raw()

	// Get task
	e.GET("/user/tasks/{task_id}", taskID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusNotFound)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
