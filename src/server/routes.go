package server

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"
	"github.com/rshelekhov/reframed/src/handlers"
	"github.com/rshelekhov/reframed/src/server/middleware/jwtoken"
	mwlogger "github.com/rshelekhov/reframed/src/server/middleware/logger"
	"time"
)

func (s *Server) initRoutes(jwtAuth *jwtoken.JWTAuth) *chi.Mux {
	r := chi.NewRouter()

	// Add request_id to each request, for tracing purposes
	r.Use(middleware.RequestID)

	// Logging of all requests
	r.Use(middleware.Logger)

	// By default, middleware.Logger uses its own src logger,
	// which should be overridden to use ours. Otherwise, problems
	// may arise - for example, with log collection. We can use
	// our own middleware to log requests:
	r.Use(mwlogger.New(s.log))

	// If a panic happens somewhere inside the server (request handlers),
	// the application should not crash.
	r.Use(middleware.Recoverer)

	// Parser of incoming request URLs
	r.Use(middleware.URLFormat)

	// Set the content type to application/json
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Enable httprate request limiter of 100 requests per minute per IP
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// Health check
	r.Get("/health", handlers.HealthRead())

	// Public routes
	r.Group(func(r chi.Router) {
		r.Post("/login", s.user.LoginWithPassword())
		r.Post("/register", s.user.CreateUser())
		// TODO: add handler for RequestResetPassword

		// Auth routes
		r.Route("/jwtoken", func(r chi.Router) {
			r.Post("/refresh-tokens", s.user.RefreshJWTTokens())
			// TODO: add handler for logout
			// r.Post("/logout", s.user.Logout())
		})
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(jwtoken.Verifier(jwtAuth))

		// Handle valid / invalid tokens
		r.Use(jwtoken.Authenticator())

		// TODO: add roles and permissions
		// Admin routes
		r.Get("/users", s.user.GetUsers())

		// User routes
		r.Route("/user", func(r chi.Router) {
			r.Route("/profile", func(r chi.Router) {
				r.Get("/", s.user.GetUser())
				r.Put("/", s.user.UpdateUser())
				r.Delete("/", s.user.DeleteUser())
			})
		})

		// list routes
		r.Route("/user/lists", func(r chi.Router) {
			r.Get("/", s.list.GetListsByUserID())
			r.Post("/", s.list.CreateList())

			r.Route("/{listID}", func(r chi.Router) {
				r.Get("/", s.list.GetListByID())
				r.Put("/", s.list.UpdateList())
				r.Delete("/", s.list.DeleteList())
			})
		})
	})

	return r
}
