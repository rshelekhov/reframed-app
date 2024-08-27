package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/segmentio/ksuid"

	"github.com/rshelekhov/reframed/internal/lib/constant/key"
	"github.com/rshelekhov/reframed/internal/model"
)

const (
	DefaultLimit = 30
)

func ParseLimitAndCursor(r *http.Request) (model.Pagination, error) {
	limit, err := strconv.Atoi(r.URL.Query().Get(key.Limit))
	if err != nil || limit < 1 {
		limit = DefaultLimit
	}

	cursor := r.URL.Query().Get(key.Cursor)

	cursorDate, err := time.Parse(time.DateOnly, cursor)
	if err == nil {
		// cursor is in date format
		return model.Pagination{
			Limit:      int32(limit),
			CursorDate: cursorDate,
		}, nil
	}

	if _, err = ksuid.Parse(cursor); err == nil {
		// cursor is in ksuid format
		return model.Pagination{
			Limit:  int32(limit),
			Cursor: cursor,
		}, nil
	}

	if cursor != "" {
		return model.Pagination{}, le.ErrInvalidCursor
	}

	// cursor is empty, it's ok
	return model.Pagination{
		Limit: int32(limit),
	}, nil
}

// ParseLimitAndAfterDate is deprecated
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
		Limit:      int32(limit),
		CursorDate: afterDate,
	}, nil
}
