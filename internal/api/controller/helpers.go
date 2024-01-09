package controller

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/go-playground/validator"
	resp "github.com/rshelekhov/reframed/internal/lib/api/response"
	"github.com/rshelekhov/reframed/pkg/logger"
	"io"
	"log/slog"
	"net/http"
)

// GetID gets the entity id from the request
func GetID(w http.ResponseWriter, r *http.Request, log logger.Interface) (string, error) {
	id := chi.URLParam(r, "id")
	if id == "" {
		log.Error("id is empty")

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, resp.Error("id is empty"))

		return "", fmt.Errorf("empty id")
	}

	return id, nil
}

// DecodeJSON decodes the request body
func DecodeJSON(w http.ResponseWriter, r *http.Request, log logger.Interface, data any) error {
	// Decode the request body
	err := render.DecodeJSON(r.Body, &data)
	if errors.Is(err, io.EOF) {
		log.Error("request body is empty")

		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, resp.Error("request body is empty"))

		return fmt.Errorf("decode error")
	}
	if err != nil {
		log.Error("failed to decode request body", logger.Err(err))

		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, resp.Error("failed to decode request body"))

		return fmt.Errorf("decode error")
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
		log.Error("failed to validate user", logger.Err(err))

		render.Status(r, http.StatusUnprocessableEntity)
		render.JSON(w, r, resp.ValidationError(ve))
	}
	if err != nil {
		log.Error("failed to validate user", logger.Err(err))

		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, resp.Error("failed to validate user"))
	}
	return nil
}
