package api_tests

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/jwtauth"
	"github.com/rshelekhov/reframed/internal/model"
)

func TestRegisterNewUser_HappyPath(t *testing.T) {
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

	// Check if access token is returned
	user.Value(jwtauth.AccessTokenKey).String().NotEmpty()

	// Check cookies
	cookie := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		Cookie(jwtauth.RefreshTokenKey)

	cookie.Value().NotEmpty()
	cookie.Domain().IsEqual(cookieDomain)
	cookie.Path().IsEqual(cookiePath)
	cookie.Expires().InRange(time.Now(), time.Now().Add(time.Hour*720))

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}

func TestRegisterNewUser_FailCases(t *testing.T) {
	testCases := []struct {
		name     string
		email    string
		password string
		status   int
	}{
		{
			name:     "Register user with empty email",
			email:    "",
			password: randomFakePassword(),
			status:   http.StatusBadRequest,
		},
		{
			name:     "Register user with empty password",
			email:    gofakeit.Email(),
			password: "",
			status:   http.StatusBadRequest,
		},
		{
			name:     "Register user with invalid email",
			email:    "invalid",
			password: randomFakePassword(),
			status:   http.StatusBadRequest,
		},
		{
			name:     "Register user when user already exists",
			email:    gofakeit.Email(),
			password: randomFakePassword(),
			status:   http.StatusConflict,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: scheme,
				Host:   host,
			}
			e := httpexpect.Default(t, u.String())

			if tc.name == "Register user when user already exists" {
				e.POST("/register").
					WithJSON(model.UserRequestData{
						Email:    tc.email,
						Password: tc.password,
					}).
					Expect().
					Status(http.StatusCreated)
			}

			e.POST("/register").
				WithJSON(model.UserRequestData{
					Email:    tc.email,
					Password: tc.password,
				}).
				Expect().
				Status(tc.status)
		})
	}
}
