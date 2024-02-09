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

	// By default, middleware.logger uses its own src logger,
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
		// TODO: add hashing password
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

		// List routes
		r.Route("/user/lists", func(r chi.Router) {
			r.Get("/", s.list.GetListsByUserID())
			r.Post("/", s.list.CreateList())

			r.Route("/{list_id}", func(r chi.Router) {
				r.Get("/", s.list.GetListByID())
				r.Put("/", s.list.UpdateList())
				r.Delete("/", s.list.DeleteList())

				r.Get("/tasks", s.task.GetTasksByListID())

				r.Route("/headings", func(r chi.Router) {
					r.Post("/", s.heading.CreateHeading())
					r.Get("/", s.heading.GetHeadingsByListID())

					r.Get("/tasks", s.task.GetTasksGroupedByHeadings())
					r.Post("/tasks", s.task.CreateTask())

					r.Route("/{heading_id}", func(r chi.Router) {
						r.Get("/", s.heading.GetHeadingByID())
						r.Put("/", s.heading.UpdateHeading())
						r.Put("/move/", s.heading.MoveHeadingToAnotherList())
						r.Delete("/", s.heading.DeleteHeading())

						r.Post("/", s.task.CreateTask())
					})
				})
			})
		})

		// Task routes
		r.Route("/user/tasks", func(r chi.Router) {
			r.Get("/", s.task.GetTasksByUserID())
			r.Get("/today", s.task.GetTasksForToday())      // grouped by list title
			r.Get("/upcoming", s.task.GetUpcomingTasks())   // grouped by start_date
			r.Get("/overdue", s.task.GetOverdueTasks())     // grouped by list title
			r.Get("/someday", s.task.GetTasksForSomeday())  // tasks without start_date, grouped by list title
			r.Get("/completed", s.task.GetCompletedTasks()) // grouped by month
			r.Get("/archived", s.task.GetArchivedTasks())   // grouped by month

			r.Route("/{task_id}", func(r chi.Router) {
				r.Get("/", s.task.GetTaskByID())
				r.Put("/", s.task.UpdateTask())
				r.Put("/time", s.task.UpdateTaskTime())
				r.Put("/move", s.task.MoveTaskToAnotherList())
				r.Put("/complete", s.task.CompleteTask())
				r.Delete("/", s.task.ArchiveTask())
			})
		})

		// Tag routes
		r.Route("/user/tags", func(r chi.Router) {
			r.Get("/", s.tag.GetTagsByUserID())
		})
	})

	return r
}
