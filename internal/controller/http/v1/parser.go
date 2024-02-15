package v1

import (
	"github.com/rshelekhov/reframed/internal/domain"
	c "github.com/rshelekhov/reframed/pkg/constants/key"
	"net/http"
	"strconv"
	"time"
)

const (
	DefaultLimit = 30
)

func ParseLimitAndAfterID(r *http.Request) domain.Pagination {
	limit, err := strconv.Atoi(r.URL.Query().Get(c.Limit))
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	afterID := r.URL.Query().Get(c.AfterID)

	return domain.Pagination{
		Limit:   limit,
		AfterID: afterID,
	}
}

func ParseLimitAndAfterDate(r *http.Request) (domain.Pagination, error) {
	limit, err := strconv.Atoi(r.URL.Query().Get(c.Limit))
	if err != nil || limit < 0 {
		limit = DefaultLimit
	}

	afterDate, err := time.Parse(time.DateOnly, r.URL.Query().Get(c.AfterDate))
	if err != nil {
		return domain.Pagination{}, err
	}

	return domain.Pagination{
		Limit:     limit,
		AfterDate: afterDate,
	}, nil
}
