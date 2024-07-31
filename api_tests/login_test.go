package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestLogin_HappyPath(t *testing.T) {
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

	// Login user and check if access token is returned
	respLogin := e.POST("/login").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	respLogin.Value(jwtoken.AccessTokenKey).String().NotEmpty()

	// Login user and check cookies
	cookie := e.POST("/login").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
		}).
		Expect().
		Status(http.StatusOK).
		Cookie(jwtoken.RefreshTokenKey)

	cookie.Value().NotEmpty()
	cookie.Domain().IsEqual(cookieDomain)
	cookie.Path().IsEqual(cookiePath)
	cookie.Expires().InRange(time.Now(), time.Now().Add(time.Hour*720))

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestLogin_FailCases(t *testing.T) {
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

	testCases := []struct {
		name     string
		email    string
		password string
		status   int
	}{
		{
			name:     "Login with empty email",
			email:    "",
			password: password,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Login with empty password",
			email:    email,
			password: "",
			status:   http.StatusBadRequest,
		},
		{
			name:     "Login with invalid email",
			email:    "invalid",
			password: password,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Login with invalid password",
			email:    email,
			password: "",
			status:   http.StatusBadRequest,
		},
		{
			name:     "Login with wrong password",
			email:    email,
			password: randomFakePassword(),
			status:   http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			e.POST("/login").
				WithJSON(model.UserRequestData{
					Email:    tc.email,
					Password: tc.password,
				}).
				Expect().
				Status(tc.status)
		})
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestLoginUserNotFound(t *testing.T) {
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

	// Delete user
	e.DELETE("/user/").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK)

	// Try to log in user
	e.POST("/login").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
		}).
		Expect().
		Status(http.StatusUnauthorized)
}
