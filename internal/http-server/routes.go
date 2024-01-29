package http_server

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"
	"github.com/rshelekhov/reframed/internal/handlers"
	"github.com/rshelekhov/reframed/internal/http-server/middleware/auth"
	mwlogger "github.com/rshelekhov/reframed/internal/http-server/middleware/logger"
	"time"
)

/*
type Router struct {
	Log  logger.Interface
	tokenAuth  *auth.JWTAuth
	user *handlers.UserHandler
	list *handlers.ListHandler
}

func NewRouter(
	log logger.Interface,
	tokenAuth *auth.JWTAuth,
	user *handlers.UserHandler,
	list *handlers.ListHandler,
) *Router {
	return &Router{
		Log:  log,
		user: user,
		list: list,
		tokenAuth:  tokenAuth,
	}
}*/

func (s *Server) initRoutes(jwtAuth *auth.JWTAuth) *chi.Mux {
	r := chi.NewRouter()

	// Add request_id to each request, for tracing purposes
	r.Use(middleware.RequestID)

	// Logging of all requests
	r.Use(middleware.Logger)

	// By default, middleware.Logger uses its own internal logger,
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
		r.Route("/auth", func(r chi.Router) {
			r.Post("/refresh-tokens", s.user.RefreshJWTTokens())
			r.Post("/logout", s.user.Logout())
		})
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		// Seek, verify and validate JWT tokens
		r.Use(auth.Verifier(jwtAuth))

		// Handle valid / invalid tokens
		r.Use(auth.Authenticator())

		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/", s.user.GetUsers())

			// TODO: use userID from JWT and remove userID from path
			r.Route("/{userID}", func(r chi.Router) {
				r.Get("/", s.user.GetUserByID())
				r.Put("/", s.user.UpdateUser())
				r.Delete("/", s.user.DeleteUser())
			})
		})

		// list routes
		r.Route("/users/{userID}/lists", func(r chi.Router) {
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
