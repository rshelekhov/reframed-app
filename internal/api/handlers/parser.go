package handlers

import (
	"github.com/rshelekhov/reframed/internal/models"
	"net/http"
	"strconv"
)

const (
	DefaultLimit  = 100
	DefaultOffset = 0
)

// ParseLimitAndOffset parses limit and offset from the request and returns a pagination object
func ParseLimitAndOffset(r *http.Request) (models.Pagination, error) {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil || offset < 0 {
		offset = DefaultOffset
	}

	pagination := models.Pagination{
		Limit:  limit,
		Offset: offset,
	}

	return pagination, nil
}
