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

func TestUpdateHeading_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
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

	listID := l.Value("data").Object().Value("id").String().Raw()

	// Create heading
	h := e.POST("/user/lists/{list_id}/headings/").
		WithPath("list_id", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	headingID := h.Value("data").Object().Value("id").String().Raw()

	// Update heading
	e.PUT("/user/lists/{list_id}/headings/{heading_id}").
		WithPath("list_id", listID).
		WithPath("heading_id", headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()
}

func TestUpdateHeading_NotFound(t *testing.T) {
	u := url.URL{
		Scheme: "http",
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

	listID := l.Value("data").Object().Value("id").String().Raw()

	// Create heading
	h := e.POST("/user/lists/{list_id}/headings/").
		WithPath("list_id", listID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: listID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	headingID := h.Value("data").Object().Value("id").String().Raw()

	// Delete heading
	e.DELETE("/user/lists/{list_id}/headings/{heading_id}").
		WithPath("list_id", listID).
		WithPath("heading_id", headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	// Update heading
	e.PUT("/user/lists/{list_id}/headings/{heading_id}").
		WithPath("list_id", listID).
		WithPath("heading_id", headingID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusNotFound)
}

func TestMoveHeadingToAnotherList_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
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

	// Create first list
	initialList := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	initialListID := initialList.Value("data").Object().Value("id").String().Raw()

	// Create second list
	otherList := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	otherListID := otherList.Value("data").Object().Value("id").String().Raw()

	// Create heading
	h := e.POST("/user/lists/{list_id}/headings/").
		WithPath("list_id", initialListID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: initialListID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	headingID := h.Value("data").Object().Value("id").String().Raw()

	// Move heading
	e.PUT("/user/lists/{list_id}/headings/{heading_id}/move").
		WithPath("list_id", initialListID).
		WithPath("heading_id", headingID).
		WithQuery("list_id", otherListID).
		WithHeader("Authorization", "Bearer "+accessToken).
		Expect().
		Status(http.StatusOK).
		JSON().Object().NotEmpty()
}

func TestMoveHadingToAnotherList_FailCases(t *testing.T) {
	u := url.URL{
		Scheme: "http",
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

	// Create first list
	initialList := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	initialListID := initialList.Value("data").Object().Value("id").String().Raw()

	// Create second list
	otherList := e.POST("/user/lists/").
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.ListRequestData{
			Title: gofakeit.Word(),
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	otherListID := otherList.Value("data").Object().Value("id").String().Raw()

	// Create heading
	h := e.POST("/user/lists/{list_id}/headings/").
		WithPath("list_id", initialListID).
		WithHeader("Authorization", "Bearer "+accessToken).
		WithJSON(model.HeadingRequestData{
			Title:  gofakeit.Word(),
			ListID: initialListID,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object()

	headingID := h.Value("data").Object().Value("id").String().Raw()

	testCases := []struct {
		name        string
		headingID   string
		otherListID string
		status      int
	}{
		{
			name:        "Move heading to another list with empty other list id",
			headingID:   headingID,
			otherListID: "",
			status:      http.StatusBadRequest,
		},
		{
			name:        "Move heading to another list with invalid other list id",
			headingID:   headingID,
			otherListID: "invalid",
			status:      http.StatusNotFound,
		},
		{
			name:        "Move heading to another list with empty heading id",
			headingID:   "",
			otherListID: otherListID,
			status:      http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e.PUT("/user/lists/{list_id}/headings/{heading_id}/move").
				WithPath("list_id", initialListID).
				WithPath("heading_id", tc.headingID).
				WithQuery("list_id", tc.otherListID).
				WithHeader("Authorization", "Bearer "+accessToken).
				Expect().
				Status(tc.status)
		})
	}
}
