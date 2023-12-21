package router

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// TODO: think to make router as a struct in the router.go
func (r *chi.Mux) routes() {

	r.Use(render.SetContentType(render.ContentTypeJSON))
}
