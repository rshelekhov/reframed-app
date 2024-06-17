package v1

import (
	"net/http"
	"strconv"
	"time"

	c "github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/model"
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

	if afterDate.IsZero() {
		afterDate = time.Now()
	}

	return model.Pagination{
		Limit:     int32(limit),
		AfterDate: afterDate,
	}, nil
}
