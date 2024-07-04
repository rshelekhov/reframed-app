package api_tests

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"net/http"
	"net/url"
	"testing"
)

func TestGetHeadingByID_HappyPath(t *testing.T) {
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
	h := e.POST("/user/lists/{list_id}/headings/", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	headingID := h.Value(key.Data).Object().Value(key.HeadingID).String().Raw()

	// Get heading
	e.GET("/user/lists/{list_id}/headings/{heading_id}", listID, headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()
}

func TestGetHeadingByID_NotFound(t *testing.T) {
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
	h := e.POST("/user/lists/{list_id}/headings/", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	headingID := h.Value(key.Data).Object().Value(key.HeadingID).String().Raw()

	// Delete heading
	e.DELETE("/user/lists/{list_id}/headings/{heading_id}", listID, headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK)

	// Get heading
	e.GET("/user/lists/{list_id}/headings/{heading_id}", listID, headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusNotFound)
}
