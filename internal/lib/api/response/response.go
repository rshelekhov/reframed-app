package response

import (
	"fmt"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
	"strings"
)

const (
	StatusOK    = "OK"
	StatusError = "ERROR"
)

type Response struct {
	Status  string    `json:"status"`
	Error   string    `json:"error,omitempty"`
	Success string    `json:"success,omitempty"`
	ID      uuid.UUID `json:"id,omitempty"`
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func Success(msg string, id uuid.UUID) Response {
	return Response{
		Status:  StatusOK,
		Success: msg,
		ID:      id,
	}
}

// ValidationError returns a Response with StatusError and a comma-separated list of errors
func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is required", err.Field()))
		case "email":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be a valid email address", err.Field()))
		case "e164":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be a valid phone number", err.Field()))
		case "min":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be greater than or equal to %s", err.Field(), err.Param()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is invalid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
