package handler

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	resp "github.com/rshelekhov/remedi/internal/lib/api/response"
	"github.com/rshelekhov/remedi/internal/lib/logger/sl"
	"io"
	"log/slog"
	"net/http"
)

// GetID gets the entity id from the request
func GetID(w http.ResponseWriter, r *http.Request, log *slog.Logger) (string, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		log.Error("id is empty")

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, resp.Error("id is empty"))

		return "", fmt.Errorf("empty id")
	}

	return id, nil
}

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

		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, resp.Error("request body is empty"))

		return fmt.Errorf("decode error")
	}
	if err != nil {
		log.Error("failed to decode request body", sl.Err(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, resp.Error("failed to decode request body"))

		return fmt.Errorf("decode error")
	}

	log.Info("request body decoded", slog.Any("user", data))

	// Validate the data
	err = v.Struct(data)
	if err != nil {
		var validateErr validator.ValidationErrors
		errors.As(err, &validateErr)

		log.Error("failed to validate user", sl.Err(err))

		render.Status(r, http.StatusUnprocessableEntity)
		render.JSON(w, r, resp.ValidationError(validateErr))

		return fmt.Errorf("validation error")
	}

	return nil
}
