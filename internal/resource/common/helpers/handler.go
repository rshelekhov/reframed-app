package helpers

import (
	"errors"
	"fmt"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	resp "github.com/rshelekhov/remedi/internal/lib/api/response"
	"github.com/rshelekhov/remedi/internal/lib/logger/sl"
	"io"
	"log/slog"
	"net/http"
)

// DecodeAndValidate decodes the request body and validates the data
func DecodeAndValidate(
	w http.ResponseWriter,
	r *http.Request,
	log *slog.Logger,
	data interface{},
	v *validator.Validate,
) error {

	// Decode the request body
	err := render.DecodeJSON(r.Body, &data)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")

		render.JSON(w, r, resp.Error(http.StatusNotFound, "request body is empty"))

		return fmt.Errorf("decode error")
	}
	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))

		render.JSON(w, r, resp.Error(http.StatusBadRequest, "failed to decode request body"))

		return fmt.Errorf("decode error")
	}

	log.Info("request body decoded", slog.Any("user", data))

	// Validate the data
	err = v.Struct(data)
	if err != nil {
		validateErr := err.(validator.ValidationErrors)

		log.Error("failed to validate user", sl.Err(err))

		render.JSON(w, r, resp.ValidationError(validateErr))

		return fmt.Errorf("validation error")
	}

	return nil
}
