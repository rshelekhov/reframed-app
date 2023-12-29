package parser

import (
	"github.com/rshelekhov/remedi/internal/model"
	"net/http"
	"strconv"
)

const (
	defaultLimit  = 100
	defaultOffset = 0
)

func ParseLimitAndOffset(r *http.Request) (model.Pagination, error) {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = defaultLimit
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		offset = defaultOffset
	}

	if limit < 0 {
		limit = defaultLimit
	}

	if offset < 0 {
		limit = defaultOffset
	}

	pagination := model.Pagination{
		Limit:  limit,
		Offset: offset,
	}

	return pagination, nil
}
