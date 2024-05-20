package api_tests

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"net/http"
	"net/url"
	"testing"
)

func TestLogout_HappyPath(t *testing.T) {
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
			AppID:    appID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := resp.Value(jwtoken.AccessTokenKey).String().Raw()
	fmt.Printf("access token: %s\n", accessToken)

	// Logout user
	e.POST("/logout").
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK)
}
