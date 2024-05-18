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
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	email := gofakeit.Email()
	password := randomFakePassword()

	// Register user
	e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
			AppID:    appID,
		}).
		Expect().
		Status(http.StatusCreated)

	// Login user and check if access token is returned
	at := e.POST("/login").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
			AppID:    appID,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	at.Value(jwtoken.AccessTokenKey).String().NotEmpty()

	// Login user and check cookies
	c := e.POST("/login").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
			AppID:    appID,
		}).
		Expect().
		Status(http.StatusOK).
		Cookie(jwtoken.RefreshTokenKey)

	c.Value().NotEmpty()
	c.Domain().IsEqual("localhost")
	c.Path().IsEqual("/")
	c.Expires().InRange(time.Now(), time.Now().Add(time.Hour*720))
}

func TestLogin_FailCases(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	email := gofakeit.Email()
	password := randomFakePassword()

	e := httpexpect.Default(t, u.String())

	// Register user
	e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
			AppID:    appID,
		}).
		Expect().
		Status(http.StatusCreated)

	testCases := []struct {
		name     string
		email    string
		password string
		appID    int32
		status   int
	}{
		{
			name:     "Login with empty email",
			email:    "",
			password: password,
			appID:    appID,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Login with empty password",
			email:    email,
			password: "",
			appID:    appID,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Login with empty app id",
			email:    email,
			password: password,
			appID:    emptyAppID,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Login with invalid email",
			email:    "invalid",
			password: password,
			appID:    appID,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Login with invalid password",
			email:    email,
			password: "",
			appID:    appID,
			status:   http.StatusBadRequest,
		},
		{
			name:     "Login with invalid app id",
			email:    email,
			password: password,
			appID:    invalidAppID,
			status:   http.StatusUnauthorized,
		},
		{
			name:     "Login with wrong password",
			email:    email,
			password: randomFakePassword(),
			appID:    appID,
			status:   http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			e.POST("/login").
				WithJSON(model.UserRequestData{
					Email:    tc.email,
					Password: tc.password,
					AppID:    tc.appID,
				}).
				Expect().
				Status(tc.status)
		})
	}
}
