package handlers

import (
	"github.com/go-chi/chi"
	"net/http"
)

func HealthHandlers(r *chi.Mux, res *Resource) {
	r.Get("/health", res.Health())
}

func (r Resource) Health() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	}
}
