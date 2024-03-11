package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rshelekhov/reframed/internal/model"
	c "github.com/rshelekhov/reframed/pkg/constants/key"
)

const (
	DefaultLimit = 30
)

func ParseLimitAndAfterID(r *http.Request) model.Pagination {
	limit, err := strconv.Atoi(r.URL.Query().Get(c.Limit))
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	afterID := r.URL.Query().Get(c.AfterID)

	return model.Pagination{
		Limit:   int32(limit),
		AfterID: afterID,
	}
}

func ParseLimitAndAfterDate(r *http.Request) (model.Pagination, error) {
	limit, err := strconv.Atoi(r.URL.Query().Get(c.Limit))
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	afterDate, err := time.Parse(time.DateOnly, r.URL.Query().Get(c.AfterDate))
	if err != nil {
		return model.Pagination{}, err
	}

	return model.Pagination{
		Limit:     int32(limit),
		AfterDate: afterDate,
	}, nil
}
