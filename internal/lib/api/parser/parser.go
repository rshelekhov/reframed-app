package parser

import (
	"github.com/go-chi/chi"
	"github.com/rshelekhov/remedi/internal/lib/api/models"
	"net/http"
	"strconv"
)

const (
	defaultLimit  = "100"
	defaultOffset = "0"
)

func ParseLimitAndOffset(r *http.Request) (models.Pagination, error) {
	limit := chi.URLParam(r, "limit")
	if limit == "" {
		limit = defaultLimit
	}

	offset := chi.URLParam(r, "offset")
	if offset == "" {
		offset = defaultOffset
	}

	limitInt, err := strconv.Atoi(limit)
	if err != nil {
		return models.Pagination{}, err
	}

	if limitInt < 0 {
		limit = defaultLimit
	}

	offsetInt, err := strconv.Atoi(offset)
	if err != nil {
		return models.Pagination{}, err
	}

	if offsetInt < 0 {
		limit = defaultOffset
	}

	pagination := models.Pagination{
		Limit:  limit,
		Offset: offset,
	}

	return pagination, nil
}
