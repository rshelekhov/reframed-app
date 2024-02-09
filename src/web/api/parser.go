package api

import (
	c "github.com/rshelekhov/reframed/src/constants"
	"github.com/rshelekhov/reframed/src/models"
	"net/http"
	"strconv"
	"time"
)

const (
	DefaultLimit = 30
)

func ParseLimitAndAfterID(r *http.Request) models.Pagination {
	limit, err := strconv.Atoi(r.URL.Query().Get(c.LimitKey))
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	afterID := r.URL.Query().Get(c.AfterIDKey)

	return models.Pagination{
		Limit:   limit,
		AfterID: afterID,
	}
}

func ParseLimitAndAfterDate(r *http.Request) (models.Pagination, error) {
	limit, err := strconv.Atoi(r.URL.Query().Get(c.LimitKey))
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	afterDate, err := time.Parse(time.DateOnly, r.URL.Query().Get(c.AfterDateKey))
	if err != nil {
		return models.Pagination{}, err
	}

	return models.Pagination{
		Limit:     limit,
		AfterDate: afterDate,
	}, nil
}
