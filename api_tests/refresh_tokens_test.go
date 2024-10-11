package api_tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/jwtauth"
	"github.com/rshelekhov/reframed/internal/model"
	"github.com/segmentio/ksuid"
)

func TestRefreshTokenUsingCookie_HappyPath(t *testing.T) {
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
		Status(http.StatusCreated)

	tokenData := user.Cookie(jwtauth.RefreshTokenKey)

	var refreshToken string
	tokenData.Value().Decode(&refreshToken)

	// Refresh tokens using cookies and check if access token is returned
	resp := e.POST("/refresh-tokens").
		WithCookie(jwtauth.RefreshTokenKey, refreshToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	resp.Value(jwtauth.AccessTokenKey).String().NotEmpty()

	// Cleanup storage after testing
	cleanupAuthService(e, user.JSON().Object())
}

func TestRefreshTokenUsingHeader_HappyPath(t *testing.T) {
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
		Status(http.StatusCreated)

	tokenData := user.Cookie(jwtauth.RefreshTokenKey)

	var refreshToken string
	tokenData.Value().Decode(&refreshToken)

	// Refresh tokens using HTTP header and check if access token is returned
	resp := e.POST("/refresh-tokens").
		WithHeader(jwtauth.RefreshTokenKey, refreshToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	resp.Value(jwtauth.AccessTokenKey).String().NotEmpty()

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user.JSON().Object())
}

func TestRefreshToken_FailCases(t *testing.T) {
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

	testCases := []struct {
		name         string
		refreshToken string
		status       int
	}{
		{
			name:         "Refresh with empty refresh token",
			refreshToken: "",
			status:       http.StatusUnauthorized,
		},
		{
			name:         "Refresh when session not found",
			refreshToken: ksuid.New().String(),
			status:       http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Refresh tokens using HTTP header
			e.POST("/refresh-tokens").
				WithCookie(jwtauth.RefreshTokenKey, tc.refreshToken).
				Expect().
				Status(tc.status)
		})
	}

	// Cleanup the SSO gRPC service storage after testing
	cleanupAuthService(e, user)
}
