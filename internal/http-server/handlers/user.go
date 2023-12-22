package handlers

import (
	"github.com/go-chi/chi"
	"net/http"
)

func UserHandlers(r *chi.Mux, res *Resource) {
	r.Get("/users", res.ListUsers())
	r.Post("/users", res.CreateUser())
	r.Get("/users/{id}", res.ReadUser())
	r.Put("/users/{id}", res.UpdateUser())
	r.Delete("/users/{id}", res.DeleteUser())
}

func (r Resource) ListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {

	}
}

func (r Resource) CreateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// user := models.CreateUser{}
	}
}

func (r Resource) ReadUser() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {

	}
}

func (r Resource) UpdateUser() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {

	}
}

func (r Resource) DeleteUser() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {

	}
}
