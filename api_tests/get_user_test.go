package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"net/http"
	"net/url"
	"testing"
)

func TestGetUser_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	resp := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := resp.Value(jwtoken.AccessTokenKey).String().Raw()

	// Get user
	e.GET("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()
}

func TestGetUser_FailCases(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	testCases := []struct {
		name        string
		accessToken string
		status      int
	}{
		{
			name:        "Get user with empty access token",
			accessToken: "",
			status:      http.StatusUnauthorized,
		},
		{
			name:        "Get user with invalid access token",
			accessToken: "invalid",
			status:      http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		e.GET("/user/").
			WithHeader("Authorization", "Bearer "+tc.accessToken).
			Expect().
			Status(tc.status)
	}
}

// TODO: add this test
// Register user and save access token to variable
// Delete user
// Try to get deleted user

//func TestGetUserNotFound(t *testing.T) {
//}
