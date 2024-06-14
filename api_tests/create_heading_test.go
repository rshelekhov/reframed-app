package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"net/http"
	"net/url"
	"testing"
)

func TestCreateHeading_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	r := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := r.Value(jwtoken.AccessTokenKey).String().Raw()

	// Create list
	l := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	listID := l.Value(key.Data).Object().Value(key.ListID).String().Raw()

	// Create heading
	e.POST("/user/lists/{list_id}/headings/", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().NotEmpty()
}

func TestCreateHeading_FailCases(t *testing.T) {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	// Register user
	r := e.POST("/register").
		WithJSON(model.UserRequestData{
			Email:    gofakeit.Email(),
			Password: randomFakePassword(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	accessToken := r.Value(jwtoken.AccessTokenKey).String().Raw()

	testCases := []struct {
		name        string
		accessToken string
		title       string
		status      int
	}{
		{
			name:        "Create heading with empty access token",
			accessToken: "",
			title:       gofakeit.Word(),
			status:      http.StatusUnauthorized,
		},
		{
			name:        "Create heading with empty title",
			accessToken: accessToken,
			title:       "",
			status:      http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create list
			c := e.POST("/user/lists/").
				WithHeader("Authorization", "Bearer "+accessToken).
				WithJSON(model.ListRequestData{
					Title: gofakeit.Word(),
				}).
				Expect().
				Status(http.StatusCreated).
				JSON().Object()

			listID := c.Value(key.Data).Object().Value(key.ListID).String().Raw()

			// Create heading
			e.POST("/user/lists/{list_id}/headings/", listID).
				WithHeader("Authorization", "Bearer "+tc.accessToken).
				WithJSON(model.HeadingRequestData{
					Title:  tc.title,
					ListID: listID,
				}).
				Expect().
				Status(tc.status)
		})
	}
}
