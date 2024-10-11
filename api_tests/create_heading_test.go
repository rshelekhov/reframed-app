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
)

func TestCreateHeading_HappyPath(t *testing.T) {
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
	e.POST("/user/lists/{list_id}/headings/", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().NotEmpty()

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestCreateHeading_FailCases(t *testing.T) {
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

	testCases := []struct {
		name        string
		accessToken string
		title       string
		status      int
	}{
		{
			name:        "Create heading with empty access token",
			accessToken: "",
			title:       gofakeit.Word(),
			status:      http.StatusUnauthorized,
		},
		{
			name:        "Create heading with empty title",
			accessToken: accessToken,
			title:       "",
			status:      http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
			e.POST("/user/lists/{list_id}/headings/", listID).
				WithHeader("Authorization", "Bearer "+tc.accessToken).
				WithJSON(model.HeadingRequestData{
					Title:  tc.title,
					ListID: listID,
				}).
				Expect().
				Status(tc.status)
		})
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestCreateHeading_InvalidUserID(t *testing.T) {
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

	accessToken1 := user1.Value(jwtauth.AccessTokenKey).String().Raw()

	// Create list
	list := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken1).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID := list.Value(key.Data).Object().Value(key.ListID).String().Raw()

	// Register second user
	user2 := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken2 := user2.Value(jwtauth.AccessTokenKey).String().Raw()

	// Create heading with invalid token and userID
	e.POST("/user/lists/{list_id}/headings/", listID).
		WithHeader("Authorization", "Bearer "+accessToken2).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object()

	// Cleanup the SSO gRPC service storage after testing
	responses := []*httpexpect.Object{user1, user2}
	for _, resp := range responses {
		cleanupAuthService(e, resp)
	}
}
