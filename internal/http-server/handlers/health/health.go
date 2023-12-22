package health

import (
	"github.com/go-chi/chi"
	"net/http"
)

func RegisterHandlers(r *chi.Mux) {
	r.Get("/health", health)
}

func health(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte("OK"))
	if err != nil {
		return
	}
}
