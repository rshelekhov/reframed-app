package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	"github.com/rshelekhov/reframed/src/le"
	"github.com/rshelekhov/reframed/src/logger"
	"io"
	"log/slog"
	"net/http"
	"reflect"
)

// ValidateData validates the request
func ValidateData(w http.ResponseWriter, r *http.Request, log logger.Interface, data any) error {
	if data == nil || reflect.DeepEqual(data, reflect.Zero(reflect.TypeOf(data)).Interface()) {
		handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyData)
		return le.ErrEmptyData
	}

	// TODO: initiate validator in the main file
	v := validator.New()
	var ve validator.ValidationErrors

	err := v.Struct(data)
	if errors.As(err, &ve) {
		log.Error(le.ErrInvalidData.Error(), logger.Err(err))
		responseValidationErrors(w, r, ve)
		return le.ErrInvalidData
	}
	if err != nil {
		handleResponseError(w, r, log, http.StatusInternalServerError, le.ErrFailedToValidateData, err)
		return le.ErrFailedToValidateData
	}
	return nil
}

// DecodeJSON decodes the request body
func DecodeJSON(w http.ResponseWriter, r *http.Request, log logger.Interface, data any) error {
	// Decode the request body
	err := render.DecodeJSON(r.Body, &data)
	if errors.Is(err, io.EOF) {
		log.Error(le.ErrEmptyRequestBody.Error())
		responseError(w, r, http.StatusBadRequest, le.ErrEmptyRequestBody)
		return le.ErrEmptyRequestBody
	}
	if err != nil {
		log.Error(le.ErrInvalidJSON.Error(), logger.Err(err))
		responseError(w, r, http.StatusBadRequest, le.ErrInvalidJSON)
		return le.ErrInvalidJSON
	}

	log.Info("request body decoded", slog.Any("user", data))

	return nil
}

func DecodeAndValidateJSON(w http.ResponseWriter, r *http.Request, log logger.Interface, data any) error {
	if err := DecodeJSON(w, r, log, data); err != nil {
		return err
	}
	if err := ValidateData(w, r, log, data); err != nil {
		return err
	}
	return nil
}

// ValidationError returns a Response with StatusError and a comma-separated list of errors
func responseValidationErrors(w http.ResponseWriter, r *http.Request, errs validator.ValidationErrors) {
	var errMessages []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMessages = append(errMessages,
				fmt.Sprintf("invalid data: field %s is required", err.Field()))
		case "email":
			errMessages = append(errMessages,
				fmt.Sprintf("invalid data: field %s must be a valid email address", err.Field()))
		case "min":
			errMessages = append(errMessages,
				fmt.Sprintf("invalid data: field %s must be greater than or equal to %s", err.Field(), err.Param()))
		}
	}

	response := struct {
		Code       int    `json:"code"`
		StatusText string `json:"status_text"`
		Data       any    `json:"data"`
	}{
		Code:       http.StatusBadRequest,
		StatusText: http.StatusText(http.StatusBadRequest),
		Data:       errMessages,
	}

	render.Status(r, http.StatusBadRequest)
	render.JSON(w, r, response)
}

func responseSuccess(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	message string,
	data any,
) {
	response := struct {
		Code        int    `json:"code"`
		StatusText  string `json:"status_text"`
		Description string `json:"description"`
		Data        any    `json:"data"`
	}{
		Code:        statusCode,
		StatusText:  http.StatusText(statusCode),
		Description: message,
		Data:        data,
	}

	render.Status(r, statusCode)
	render.JSON(w, r, response)
}

// responseError renders an error response with the given status code and error
func responseError(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	errorMessage le.LocalError,
) {
	response := struct {
		Code        int    `json:"code"`
		StatusText  string `json:"status_text"`
		Description string `json:"description"`
	}{
		Code:        statusCode,
		StatusText:  http.StatusText(statusCode),
		Description: fmt.Sprintf("%v", errorMessage),
	}

	render.Status(r, statusCode)
	render.JSON(w, r, response)
}

// handleResponseError renders an error response with the given status code and error
func handleResponseError(
	w http.ResponseWriter,
	r *http.Request,
	log logger.Interface,
	status int,
	error le.LocalError,
	additionalData ...interface{},
) {
	log.Error(fmt.Sprintf("%v", error), additionalData...)
	responseError(w, r, status, error)
}

func handleInternalServerError(
	w http.ResponseWriter,
	r *http.Request,
	log logger.Interface,
	error le.LocalError,
	additionalData ...interface{},
) {
	log.Error(fmt.Sprintf("%v", error), additionalData...)
	responseError(w, r, http.StatusInternalServerError, error)
}
