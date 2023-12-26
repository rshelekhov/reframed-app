package parser

import (
	"github.com/go-chi/chi"
	"github.com/rshelekhov/remedi/internal/lib/api/models"
	"net/http"
	"strconv"
)

const (
	defaultLimit  = 100
	defaultOffset = 0
)

func ParseLimitAndOffset(r *http.Request) (models.Pagination, error) {
	limit, err := strconv.Atoi(chi.URLParam(r, "limit"))
	if err != nil {
		limit = defaultLimit
	}

	offset, err := strconv.Atoi(chi.URLParam(r, "offset"))
	if err != nil {
		offset = defaultOffset
	}

	if limit < 0 {
		limit = defaultLimit
	}

	pagination := models.Pagination{
		Limit:  limit,
		Offset: offset,
	}

	return pagination, nil
}
