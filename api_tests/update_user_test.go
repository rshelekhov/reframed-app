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

func TestUpdateUser_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
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

	// Update user
	e.PUT("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusOK)
}

func TestUpdateUser_FailCases(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	email := gofakeit.Email()
	password := randomFakePassword()

	// Register user
	resp := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := resp.Value(jwtoken.AccessTokenKey).String().Raw()

	testCases := []struct {
		name        string
		email       string
		curPassword string
		updPassword string
		status      int
	}{
		{
			name:        "Update user when no email changes detected",
			email:       email,
			curPassword: password,
			updPassword: randomFakePassword(),
			status:      http.StatusBadRequest,
		},
		{
			name:        "Update user when no password changes detected",
			email:       gofakeit.Email(),
			curPassword: password,
			updPassword: password,
			status:      http.StatusBadRequest,
		},
		{
			name:        "Update user when current password is incorrect",
			email:       gofakeit.Email(),
			curPassword: randomFakePassword(),
			updPassword: randomFakePassword(),
			status:      http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Update user
			e.PUT("/user/").
				WithHeader("Authorization", "Bearer "+accessToken).
				WithJSON(model.UserRequestData{
					Email:           tc.email,
					Password:        tc.curPassword,
					UpdatedPassword: tc.updPassword,
				}).
				Expect().
				Status(tc.status)
		})
	}
}

func TestUpdateUserNotFound(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
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

	// Delete user
	e.DELETE("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK)

	// Try to update user
	e.PUT("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusOK)
	//	status: http.StatusNotFound,
}

func TestUpdateUserEmailAlreadyTaken(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	email := gofakeit.Email()

	// Register first user
	resp := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := resp.Value(jwtoken.AccessTokenKey).String().Raw()

	// Register second user
	e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated)

	// Try to update user
	e.PUT("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusConflict)
}
