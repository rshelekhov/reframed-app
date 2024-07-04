package v1

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/rshelekhov/reframed/internal/lib/constant/le"
	"github.com/rshelekhov/reframed/internal/lib/logger"
	"log/slog"
	"net/http"
	"reflect"
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
