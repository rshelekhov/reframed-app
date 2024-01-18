package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/rshelekhov/reframed/internal/logger"
	"io"
	"log/slog"
	"net/http"
)

// GetID gets the models id from the request
func GetID(w http.ResponseWriter, r *http.Request, log logger.Interface) (string, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		log.Error("id is empty")
		responseError(w, r, http.StatusBadRequest, "id is empty")
		return "", fmt.Errorf("empty id")
	}

	return id, nil
}

type RequestDecoder interface {
	decodeJSON(w http.ResponseWriter, r *http.Request, log logger.Interface, data any) error
}

// decodeJSON decodes the request body
func decodeJSON(w http.ResponseWriter, r *http.Request, log logger.Interface, data any) error {
	// Decode the request body
	err := render.DecodeJSON(r.Body, &data)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")
		responseError(w, r, http.StatusBadRequest, "request body is empty")
		return fmt.Errorf("decode error")
	}
	if err != nil {
		log.Error("failed to decode request body", logger.Err(err))
		responseError(w, r, http.StatusBadRequest, "failed to decode request body")
		return fmt.Errorf("decode error")
	}

	log.Info("request body decoded", slog.Any("user", data))

	return nil
}

// validateData validates the request
func validateData(w http.ResponseWriter, r *http.Request, log logger.Interface, data any) error {
	// TODO: initiate validator in the main file
	v := validator.New()
	var ve validator.ValidationErrors

	err := v.Struct(data)
	if errors.As(err, &ve) {
		log.Error("failed to validate user", logger.Err(err))
		responseValidationErrors(w, r, ve)
		return fmt.Errorf("validation error")
	}
	if err != nil {
		log.Error("failed to validate user", logger.Err(err))
		responseError(w, r, http.StatusInternalServerError, "failed to validate user")
		return fmt.Errorf("validation error")
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
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is invalid", err.Field()))
		}
	}

	response := struct {
		Code       int    `json:"code"`
		StatusText string `json:"status_text"`
		Data       any    `json:"data"`
	}{
		Code:       http.StatusUnprocessableEntity,
		StatusText: http.StatusText(http.StatusUnprocessableEntity),
		Data:       errMsgs,
	}

	render.Status(r, http.StatusUnprocessableEntity)
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
