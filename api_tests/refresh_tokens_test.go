package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/model"
	"net/http"
	"net/url"
	"testing"
)

func TestRefreshTokenUsingCookie_HappyPath(t *testing.T) {

	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	tokenData := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
			AppID:    appID,
		}).
		Expect().
		Status(http.StatusCreated).
		Cookie("refreshToken")

	var refreshToken string
	tokenData.Value().Decode(&refreshToken)

	// Refresh tokens using cookies and check if access token is returned
	c := e.POST("/refresh-tokens").
		WithCookie("refreshToken", refreshToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	c.Value("accessToken").String().NotEmpty()
}

func TestRefreshTokenUsingHeader_HappyPath(t *testing.T) {

	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	tokenData := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
			AppID:    appID,
		}).
		Expect().
		Status(http.StatusCreated).
		Cookie("refreshToken")

	var refreshToken string
	tokenData.Value().Decode(&refreshToken)

	// Refresh tokens using HTTP header and check if access token is returned
	h := e.POST("/refresh-tokens").
		WithHeader("refreshToken", refreshToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	h.Value("accessToken").String().NotEmpty()
}
