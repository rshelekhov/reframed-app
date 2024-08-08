package v1

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	"github.com/rshelekhov/reframed/internal/model"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"
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

// getUserIDFromContext retrieves the userID from context and handles any errors.
func getUserIDFromContext(ctx context.Context, w http.ResponseWriter, r *http.Request, jwt *jwtoken.TokenService, log *slog.Logger) (userID string, err error) {
	userID, err = jwt.GetUserID(ctx)
	switch {
	case errors.Is(err, jwtoken.ErrUserIDNotFoundInCtx):
		handleResponseError(w, r, log, http.StatusNotFound, le.LocalError(jwtoken.ErrUserIDNotFoundInCtx.Error()))
		return "", err
	case err != nil:
		handleInternalServerError(w, r, log, le.ErrFailedToGetUserIDFromToken, err)
		return "", err
	}
	return userID, nil
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

type errorResponse struct {
	Error      le.LocalError `json:"error"`
	StatusCode int           `json:"status_code"`
	Location   string        `json:"location"`
	Time       time.Time     `json:"time"`
}

// responseError renders an error response with the given status code and error
func responseError(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	localError le.LocalError,
	resp *errorResponse,
) {
	_, file, line, ok := getCaller()
	if !ok {
		file = "unknown file"
		line = -1
	}

	// op := runtime.FuncForPC(pc).Name()
	location := fmt.Sprintf("%s:%d", file, line)

	resp.Error = localError
	resp.StatusCode = status
	// resp.Operation = op
	resp.Location = location
	resp.Time = time.Now()

	render.Status(r, status)
	render.JSON(w, r, resp)
}

func getCaller() (pc uintptr, file string, line int, ok bool) {
	i := 1
	for {
		p, f, l, o := runtime.Caller(i)
		if !o {
			return
		}
		fmt.Println(f)
		if strings.Contains(f, "reframed") {
			pc = p
			file = f
			line = l
			ok = o
		} else {
			return
		}
		i += 1
	}
}

// handleResponseError renders an error response with the given status code and error
func handleResponseError(
	w http.ResponseWriter,
	r *http.Request,
	log *slog.Logger,
	status int,
	err le.LocalError,
	addLogData ...interface{},
) {
	resp := &errorResponse{}
	responseError(w, r, status, err, resp)
	log.Error(fmt.Sprintf("err: %v (status %v). Where: %v", err, status, resp.Location), addLogData...)
}

func handleInternalServerError(
	w http.ResponseWriter,
	r *http.Request,
	log *slog.Logger,
	err le.LocalError,
	errDetails error,
) {
	resp := &errorResponse{}
	responseError(w, r, http.StatusInternalServerError, err, resp)
	log.Error(fmt.Sprintf("error: %v (status %v). Details: %s.Where: %v", err, http.StatusInternalServerError, errDetails, resp.Location))
}
