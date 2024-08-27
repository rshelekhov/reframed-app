package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/model"
)

func decodeJSON(w http.ResponseWriter, r *http.Request, log *slog.Logger, data any) error {
	var err error
	switch v := data.(type) {
	case *model.TaskRequestData:
		err = decodeTaskRequestData(r, v)
	case *model.TaskRequestTimeData:
		err = decodeTaskRequestTimeData(r, v)
	default:
		err = render.DecodeJSON(r.Body, &data)
	}

	if errors.Is(err, io.EOF) {
		log.Error(le.ErrEmptyRequestBody.Error())

		resp := &errorResponse{}
		responseError(w, r, http.StatusBadRequest, le.ErrEmptyRequestBody, resp)

		return le.ErrEmptyRequestBody
	}
	if err != nil {
		log.Error(le.ErrInvalidJSON.Error(), logger.Err(err))

		resp := &errorResponse{}
		responseError(w, r, http.StatusBadRequest, le.LocalError(err.Error()), resp)

		return le.ErrInvalidJSON
	}

	log.Info("request body decoded", slog.Any("user", data))

	return nil
}

func decodeTaskRequestData(r *http.Request, data *model.TaskRequestData) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	// Manually parse time fields
	var err error

	data.StartDateParsed, err = parseIfNotEmpty(data.StartDate, func(v string) (time.Time, error) {
		return time.Parse(time.DateOnly, v)
	})
	if err != nil {
		return fmt.Errorf("invalid start_date format, got %s, need to use the following format: %s", data.StartDate, time.DateOnly)
	}

	data.DeadlineParsed, err = parseIfNotEmpty(data.Deadline, func(v string) (time.Time, error) {
		return time.Parse(time.DateOnly, v)
	})
	if err != nil {
		return fmt.Errorf("invalid deadline format, got %s, need to use the following format: %s", data.StartDate, time.DateOnly)
	}

	data.StartTimeParsed, err = parseIfNotEmpty(data.StartTime, func(v string) (time.Time, error) {
		return time.Parse(time.DateTime, v)
	})
	if err != nil {
		return fmt.Errorf("invalid start_time format, got %s, need to use the following format: %s", data.StartDate, time.DateTime)
	}

	data.EndTimeParsed, err = parseIfNotEmpty(data.EndTime, func(v string) (time.Time, error) {
		return time.Parse(time.DateTime, v)
	})
	if err != nil {
		return fmt.Errorf("invalid end_time forma, got %s, need to use the following format: %s", data.StartDate, time.DateTime)
	}

	return nil
}

func decodeTaskRequestTimeData(r *http.Request, data *model.TaskRequestTimeData) error {
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	// Manually parse time fields
	var err error

	data.StartTimeParsed, err = parseIfNotEmpty(data.StartTime, func(v string) (time.Time, error) {
		return time.Parse(time.DateTime, v)
	})
	if err != nil {
		return fmt.Errorf("invalid start_time format, got %s, need to use the following format: %s", data.StartTime, time.DateTime)
	}

	data.EndTimeParsed, err = parseIfNotEmpty(data.EndTime, func(v string) (time.Time, error) {
		return time.Parse(time.DateTime, v)
	})
	if err != nil {
		return fmt.Errorf("invalid end_time forma, got %s, need to use the following format: %s", data.EndTime, time.DateTime)
	}

	return nil
}

func parseIfNotEmpty(value string, parseFunc func(string) (time.Time, error)) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return parseFunc(value)
}
