package handlers_test

import (
	"github.com/rshelekhov/reframed/src/handlers"
	"github.com/rshelekhov/reframed/src/models"
	"net/http"
	"net/url"
	"testing"
)

func TestParseLimitAndOffset(t *testing.T) {
	testCases := []struct {
		name       string
		url        string
		pagination models.Pagination
	}{
		{
			name:       "valid limit and offset",
			url:        "https://example.com?limit=10&offset=5",
			pagination: models.Pagination{Limit: 10, Offset: 5},
		},
		{
			name:       "limit and offset not provided",
			url:        "https://example.com",
			pagination: models.Pagination{Limit: handlers.DefaultLimit, Offset: handlers.DefaultOffset},
		},
		{
			name:       "invalid limit or offset",
			url:        "https://example.com?limit=abc&offset=xyz",
			pagination: models.Pagination{Limit: handlers.DefaultLimit, Offset: handlers.DefaultOffset},
		},
		{
			name:       "negative limit and offset",
			url:        "https://example.com?limit=-1&offset=-1",
			pagination: models.Pagination{Limit: handlers.DefaultLimit, Offset: handlers.DefaultOffset},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			url, _ := url.Parse(tc.url)
			req := &http.Request{URL: url}

			pagination, err := handlers.ParseLimitAndOffset(req)
			if err != nil {
				t.Error(err)
			}
			if pagination != tc.pagination {
				t.Errorf("Expected: %v, but got: %v", tc.pagination, pagination)
			}
		})
	}
}
