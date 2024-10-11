package api_tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/jwtauth"
	"github.com/rshelekhov/reframed/internal/model"
)

func TestCreateList_HappyPath(t *testing.T) {
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
	e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().NotEmpty()

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestCreateList_FailCases(t *testing.T) {
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
			name:        "Create list with empty access token",
			accessToken: "",
			title:       gofakeit.Word(),
			status:      http.StatusUnauthorized,
		},
		{
			name:        "Create list with empty title",
			accessToken: accessToken,
			title:       "",
			status:      http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e.POST("/user/lists/").
				WithHeader("Authorization", "Bearer "+tc.accessToken).
				WithJSON(model.ListRequestData{
					Title: tc.title,
				}).
				Expect().
				Status(tc.status)
		})
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
