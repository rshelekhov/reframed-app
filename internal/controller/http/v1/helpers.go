package v1

import (
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/rshelekhov/reframed/internal/lib/constants/le"
	"github.com/rshelekhov/reframed/internal/model"
	"log/slog"
	"net/http"
)

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
	response := model.Response{
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
	addLogData ...interface{}, // TODO: use map instead (avoid !BADKEY in logs)
) {
	log.Error("Internal Server Error: ", addLogData...)
	responseError(w, r, http.StatusInternalServerError, error)
}
