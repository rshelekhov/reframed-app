package v1

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
)

// validateData validates the request
func validateData(w http.ResponseWriter, r *http.Request, log *slog.Logger, data any) error {
	if data == nil || reflect.DeepEqual(data, reflect.Zero(reflect.TypeOf(data)).Interface()) {
		handleResponseError(w, r, log, http.StatusBadRequest, le.ErrEmptyData)
		return le.ErrEmptyData
	}

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

// decodeJSON decodes the request body
func decodeJSON(w http.ResponseWriter, r *http.Request, log *slog.Logger, data any) error {
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

func decodeAndValidateJSON(w http.ResponseWriter, r *http.Request, log *slog.Logger, data any) error {
	if err := decodeJSON(w, r, log, data); err != nil {
		return err
	}
	if err := validateData(w, r, log, data); err != nil {
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

// handleResponseSuccess renders a success response with status code and data
func handleResponseSuccess(
	w http.ResponseWriter,
	r *http.Request,
	log *slog.Logger,
	message string,
	data any,
	addLogData ...any,
) {
	log.Info(message, addLogData...)
	responseSuccess(w, r, http.StatusOK, message, data)
}

// handleResponseCreated renders a created response with status code and data
func handleResponseCreated(
	w http.ResponseWriter,
	r *http.Request,
	log *slog.Logger,
	message string,
	data any,
	addLogData ...any,
) {
	log.Info(message, addLogData...)
	responseSuccess(w, r, http.StatusCreated, message, data)
}

// responseError renders an error response with the given status code and error
func responseError(
	w http.ResponseWriter,
	r *http.Request,
	statusCode int,
	errorMessage le.LocalError,
) {
	errorResponse := struct {
		Code        int    `json:"code"`
		StatusText  string `json:"status_text"`
		Description string `json:"description"`
	}{
		Code:        statusCode,
		StatusText:  http.StatusText(statusCode),
		Description: fmt.Sprintf("%v", errorMessage),
	}

	render.Status(r, statusCode)
	render.JSON(w, r, errorResponse)
}

// handleResponseError renders an error response with the given status code and error
func handleResponseError(
	w http.ResponseWriter,
	r *http.Request,
	log *slog.Logger,
	status int,
	error le.LocalError,
	addLogData ...interface{},
) {
	log.Error(fmt.Sprintf("%v", error), addLogData...)
	responseError(w, r, status, error)
}

func handleInternalServerError(
	w http.ResponseWriter,
	r *http.Request,
	log *slog.Logger,
	error le.LocalError,
	addLogData ...interface{},
) {
	log.Error("Internal Server Error: ", addLogData...)
	responseError(w, r, http.StatusInternalServerError, error)
}
