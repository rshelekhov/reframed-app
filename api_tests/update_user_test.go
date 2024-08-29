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

	pass := randomFakePassword()

	// Register user
	user := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: pass,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := user.Value(jwtoken.AccessTokenKey).String().Raw()

	// Update user
	e.PATCH("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.UpdateUserRequestData{
			Email:           gofakeit.Email(),
			Password:        pass,
			UpdatedPassword: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusOK)

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
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
	user := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := user.Value(jwtoken.AccessTokenKey).String().Raw()

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
			e.PATCH("/user/").
				WithHeader("Authorization", "Bearer "+accessToken).
				WithJSON(model.UpdateUserRequestData{
					Email:           tc.email,
					Password:        tc.curPassword,
					UpdatedPassword: tc.updPassword,
				}).
				Expect().
				Status(tc.status)
		})
	}

	// Clearing the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestUpdateUserNotFound(t *testing.T) {
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

	// Delete user
	e.DELETE("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK)

	// Try to update user
	e.PATCH("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusNotFound)
}

func TestUpdateUserEmailAlreadyTaken(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	email := gofakeit.Email()

	// Register first user
	user1 := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := user1.Value(jwtoken.AccessTokenKey).String().Raw()

	// Register second user
	user2 := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	// Try to update user
	e.PATCH("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusConflict)

	// Cleanup the SSO gRPC service storage after testing
	responses := []*httpexpect.Object{user1, user2}
	for _, resp := range responses {
		cleanupAuthService(e, resp)
	}
}
