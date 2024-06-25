package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rshelekhov/reframed/internal/lib/constants/key"
	"github.com/rshelekhov/reframed/internal/model"
)

const (
	DefaultLimit = 30
)

func ParseLimitAndAfterID(r *http.Request) model.Pagination {
	limit, err := strconv.Atoi(r.URL.Query().Get(key.Limit))
	if err != nil || limit < 1 {
		limit = DefaultLimit
	}

	afterID := r.URL.Query().Get(key.AfterID)

	return model.Pagination{
		Limit:   int32(limit),
		AfterID: afterID,
	}
}

func ParseLimitAndAfterDate(r *http.Request) (model.Pagination, error) {
	limit, err := strconv.Atoi(r.URL.Query().Get(key.Limit))
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	var afterDate time.Time
	afterDateString := r.URL.Query().Get(key.AfterDate)

	if afterDateString == "" {
		afterDate = time.Now()
	} else {
		afterDate, err = time.Parse(time.DateOnly, afterDateString)
		if err != nil {
			return model.Pagination{}, err
		}
	}

	return model.Pagination{
		Limit:     int32(limit),
		AfterDate: afterDate,
	}, nil
}
