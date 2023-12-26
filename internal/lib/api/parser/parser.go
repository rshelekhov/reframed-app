package parser

import (
	"github.com/rshelekhov/remedi/internal/lib/api/models"
	"log/slog"
	"net/http"
	"strconv"
)

const (
	defaultLimit  = 100
	defaultOffset = 0
)

func ParseLimitAndOffset(log *slog.Logger, r *http.Request) (models.Pagination, error) {
	var limit, offset int

	query := r.URL.Query()

	limitStr := query.Get("limit")
	if limitStr == "" {
		limit = defaultLimit
		log.Info("limit not specified, using default value for limit", slog.Int("limit", limit))
	} else {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			return models.Pagination{}, err
		}
	}

	offsetStr := query.Get("offset")
	if offsetStr == "" {
		offset = defaultOffset
		log.Info("offset not specified, using default value for offset", slog.Int("offset", offset))
	} else {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return models.Pagination{}, err
		}
	}

	if limit < 0 {
		limit = defaultLimit
		log.Info("limit cannot be negative, using default value for limit", slog.Int("limit", limit))
	}

	if offset < 0 {
		limit = defaultOffset
		log.Info("offset cannot be negative, using default value offset", slog.Int("offset", offset))
	}

	pagination := models.Pagination{
		Limit:  limit,
		Offset: offset,
	}

	return pagination, nil
}
