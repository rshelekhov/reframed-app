package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"github.com/rshelekhov/reframed/internal/model"
	"io"
	"log/slog"
	"net/http"
	"time"
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
		responseError(w, r, http.StatusBadRequest, le.ErrEmptyRequestBody)
		return le.ErrEmptyRequestBody
	}
	if err != nil {
		log.Error(le.ErrInvalidJSON.Error(), logger.Err(err))
		responseError(w, r, http.StatusBadRequest, le.LocalError(err.Error()))
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
		return errors.New("invalid deadline format")
	}

	data.StartTimeParsed, err = parseIfNotEmpty(data.StartTime, func(v string) (time.Time, error) {
		return time.Parse(time.DateTime, v)
	})
	if err != nil {
		return errors.New("invalid start_time format")
	}

	data.EndTimeParsed, err = parseIfNotEmpty(data.EndTime, func(v string) (time.Time, error) {
		return time.Parse(time.DateTime, v)
	})
	if err != nil {
		return errors.New("invalid end_time format")
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
		return errors.New("invalid start_time format")
	}

	data.EndTimeParsed, err = parseIfNotEmpty(data.EndTime, func(v string) (time.Time, error) {
		return time.Parse(time.DateTime, v)
	})
	if err != nil {
		return errors.New("invalid end_time format")
	}

	return nil
}

func parseIfNotEmpty(value string, parseFunc func(string) (time.Time, error)) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	return parseFunc(value)
}
