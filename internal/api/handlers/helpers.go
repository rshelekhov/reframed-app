package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/rshelekhov/reframed/internal/logger"
	"io"
	"log/slog"
	"net/http"
)

var (
	ErrEmptyID = "id is empty"

	ErrEmptyRequestBody = "request body is empty"
	ErrInvalidJSON      = "failed to decode request body"

	ErrInvalidData = "failed to validate data"
)

// GetID gets the models id from the request
func GetID(w http.ResponseWriter, r *http.Request, log logger.Interface) (string, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		log.Error(ErrEmptyID)
		responseError(w, r, http.StatusBadRequest, ErrEmptyID)
		return "", fmt.Errorf(ErrEmptyID)
	}

	return id, nil
}

// DecodeJSON decodes the request body
func DecodeJSON(w http.ResponseWriter, r *http.Request, log logger.Interface, data any) error {
	// Decode the request body
	err := render.DecodeJSON(r.Body, &data)
	if errors.Is(err, io.EOF) {
		log.Error(ErrEmptyRequestBody)
		responseError(w, r, http.StatusBadRequest, ErrEmptyRequestBody)
		return fmt.Errorf(ErrEmptyRequestBody)
	}
	if err != nil {
		log.Error(ErrInvalidJSON, logger.Err(err))
		responseError(w, r, http.StatusBadRequest, ErrInvalidJSON)
		return fmt.Errorf(ErrInvalidJSON)
	}

	log.Info("request body decoded", slog.Any("user", data))

	return nil
}

// ValidateData validates the request
func ValidateData(w http.ResponseWriter, r *http.Request, log logger.Interface, data any) error {
	// TODO: initiate validator in the main file
	v := validator.New()
	var ve validator.ValidationErrors

	err := v.Struct(data)
	if errors.As(err, &ve) {
		log.Error(ErrInvalidData, logger.Err(err))
		responseValidationErrors(w, r, ve)
		return fmt.Errorf(ErrInvalidData)
	}
	if err != nil {
		log.Error(ErrInvalidData, logger.Err(err))
		responseError(w, r, http.StatusInternalServerError, ErrInvalidData)
		return fmt.Errorf(ErrInvalidData)
	}
	return nil
}

// ValidationError returns a Response with StatusError and a comma-separated list of errors
func responseValidationErrors(w http.ResponseWriter, r *http.Request, errs validator.ValidationErrors) {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is required", err.Field()))
		case "email":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be a valid email address", err.Field()))
		case "min":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be greater than or equal to %s", err.Field(), err.Param()))
		}
	}

	response := struct {
		Code       int    `json:"code"`
		StatusText string `json:"status_text"`
		Data       any    `json:"data"`
	}{
		Code:       http.StatusBadRequest,
		StatusText: http.StatusText(http.StatusBadRequest),
		Data:       errMsgs,
	}

	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, response)
}

func responseSuccess(w http.ResponseWriter, r *http.Request, statusCode int, msg string, data any) {
	response := struct {
		Code        int    `json:"code"`
		StatusText  string `json:"status_text"`
		Description string `json:"description"`
		Data        any    `json:"data"`
	}{
		Code:        statusCode,
		StatusText:  http.StatusText(statusCode),
		Description: msg,
		Data:        data,
	}

	render.Status(r, statusCode)
	render.JSON(w, r, response)
}

// responseError renders an error response with the given status code and error
func responseError(w http.ResponseWriter, r *http.Request, statusCode int, msg string) {
	response := struct {
		Code        int    `json:"code"`
		StatusText  string `json:"status_text"`
		Description string `json:"description"`
	}{
		Code:        statusCode,
		StatusText:  http.StatusText(statusCode),
		Description: msg,
	}

	render.Status(r, statusCode)
	render.JSON(w, r, response)
}
