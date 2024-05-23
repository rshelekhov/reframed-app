package v1

import (
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"
	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"
	mwlogger "github.com/rshelekhov/reframed/internal/lib/middleware/logger"
	"time"
)

func (ar *AppRouter) initRoutes() *chi.Mux {
	r := chi.NewRouter()

	// Strip trailing slashes
	r.Use(middleware.StripSlashes)

	// Add request_id to each request, for tracing purposes
	r.Use(middleware.RequestID)

	// Logging of all requests
	r.Use(middleware.Logger)

	// By default, middleware.logger uses its own internal logger,
	// which should be overridden to use ours. Otherwise, problems
	// may arise - for example, with log collection. We can use
	// our own middleware to log requests:
	r.Use(mwlogger.New(ar.Logger))

	// If a panic happens somewhere inside the httpserver (request controller),
	// the application should not crash.
	r.Use(middleware.Recoverer)

	// Parser of incoming request URLs
	r.Use(middleware.URLFormat)

	// Set the content type to application/json
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// Enable httprate request limiter of 100 requests per minute per IP
	r.Use(httprate.LimitByIP(100, 1*time.Minute))

	// Health check
	r.Get("/health", HealthRead())

	//Public routes
	r.Group(func(r chi.Router) {
		r.Post("/login", ar.LoginWithPassword())
		r.Post("/register", ar.Register())
		// TODO: add handler for RequestResetPassword
		r.Post("/refresh-tokens", ar.RefreshJWTokens())
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(jwtoken.Verifier(ar.TokenService))
		r.Use(jwtoken.Authenticator())

		r.Route("/user", func(r chi.Router) {
			r.Get("/", ar.GetUser())
			r.Put("/", ar.UpdateUser())
			r.Delete("/", ar.DeleteUser())

			r.Route("/lists", func(r chi.Router) {
				r.Get("/", ar.GetListsByUserID())
				r.Post("/", ar.CreateList())
				// TODO: Add handler for creating task in the inbox list
				r.Post("/default", ar.CreateTaskInDefaultList())

				r.Route("/{list_id}", func(r chi.Router) {
					r.Get("/", ar.GetListByID())
					r.Put("/", ar.UpdateList())
					r.Delete("/", ar.DeleteList())

					r.Route("/tasks", func(r chi.Router) {
						r.Get("/", ar.GetTasksByListID())
						r.Post("/", ar.CreateTask())
					})

					r.Route("/headings", func(r chi.Router) {
						r.Post("/", ar.CreateHeading())
						r.Get("/", ar.GetHeadingsByListID())
						r.Get("/tasks", ar.GetTasksGroupedByHeadings())
						r.Post("/{heading_id}", ar.CreateTask())

						r.Route("/{heading_id}", func(r chi.Router) {
							r.Get("/", ar.GetHeadingByID())
							r.Put("/", ar.UpdateHeading())
							r.Put("/move/", ar.MoveHeadingToAnotherList())
							r.Delete("/", ar.DeleteHeading())
						})
					})
				})
			})

			r.Route("/tasks", func(r chi.Router) {
				r.Get("/", ar.GetTasksByUserID())
				r.Get("/today", ar.GetTasksForToday())      // grouped by list title
				r.Get("/upcoming", ar.GetUpcomingTasks())   // grouped by start_date
				r.Get("/overdue", ar.GetOverdueTasks())     // grouped by list title
				r.Get("/someday", ar.GetTasksForSomeday())  // tasks without start_date, grouped by list title
				r.Get("/completed", ar.GetCompletedTasks()) // grouped by month
				r.Get("/archived", ar.GetArchivedTasks())   // grouped by month

				r.Route("/{task_id}", func(r chi.Router) {
					r.Get("/", ar.GetTaskByID())
					r.Put("/", ar.UpdateTask())
					r.Put("/time", ar.UpdateTaskTime())
					r.Put("/move", ar.MoveTaskToAnotherList())
					r.Put("/complete", ar.CompleteTask())
					r.Delete("/", ar.ArchiveTask())
				})
			})

			r.Get("/tags", ar.GetTagsByUserID())
		})
	})

	return r
}
