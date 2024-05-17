package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/model"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestLoginUser_HappyPath(t *testing.T) {
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

	at.Value("accessToken").String().NotEmpty()

	// Login user and check cookies
	c := e.POST("/login").
		WithJSON(model.UserRequestData{
			Email:    email,
			Password: password,
			AppID:    appID,
		}).
		Expect().
		Status(http.StatusOK).
		Cookie("refreshToken")

	c.Value().NotEmpty()
	c.Domain().IsEqual("localhost")
	c.Path().IsEqual("/")
	c.Expires().InRange(time.Now(), time.Now().Add(time.Hour*720))
}
