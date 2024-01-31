package handlers

import (
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/models"
	"net/http"
	"strconv"
)

const (
	DefaultLimit  = 100
	DefaultOffset = 0
)

// ParseLimitAndOffset parses limit and offset from the request and returns a pagination object
func ParseLimitAndOffset(r *http.Request) (models.Pagination, error) {
	limit, err := strconv.Atoi(r.URL.Query().Get(c.LimitKey))
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	offset, err := strconv.Atoi(r.URL.Query().Get(c.OffsetKey))
	if err != nil || offset < 0 {
		offset = DefaultOffset
	}

	pagination := models.Pagination{
		Limit:  limit,
		Offset: offset,
	}

	return pagination, nil
}
