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
	r.Use(httprate.LimitByIP(ar.ServerSettings.HTTPServer.RequestLimitByIP, 1*time.Minute))

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

		r.Post("/logout", ar.Logout())

		r.Route("/statuses", func(r chi.Router) {
			r.Get("/", ar.GetStatuses())
			r.Get("/{status_id}", ar.GetStatusByID())
		})

		r.Route("/user", func(r chi.Router) {
			r.Get("/", ar.GetUser())
			r.Patch("/", ar.UpdateUser())
			r.Delete("/", ar.DeleteUser())

			r.Route("/lists", func(r chi.Router) {
				r.Get("/", ar.GetListsByUserID())
				r.Post("/", ar.CreateList())
				r.Get("/default", ar.GetDefaultList())
				r.Post("/default", ar.CreateTaskInDefaultList())

				r.Route("/{list_id}", func(r chi.Router) {
					r.Get("/", ar.GetListByID())
					r.Patch("/", ar.UpdateList())
					r.Delete("/", ar.DeleteList())

					r.Route("/tasks", func(r chi.Router) {
						r.Get("/", ar.GetTasksByListID())
						r.Post("/", ar.CreateTask())
					})

					r.Route("/headings", func(r chi.Router) {
						r.Post("/", ar.CreateHeading())
						r.Get("/", ar.GetHeadingsByListID())
						r.Get("/tasks", ar.GetTasksGroupedByHeadings())

						r.Route("/{heading_id}", func(r chi.Router) {
							r.Post("/", ar.CreateTask())
							r.Get("/", ar.GetHeadingByID())
							r.Patch("/", ar.UpdateHeading())
							r.Patch("/move", ar.MoveHeadingToAnotherList())
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

				// TODO: replace Put with Patch (here and for other handlers for updating data)
				r.Route("/{task_id}", func(r chi.Router) {
					r.Get("/", ar.GetTaskByID())
					r.Patch("/", ar.UpdateTask())
					r.Patch("/time", ar.UpdateTaskTime())
					r.Patch("/move/list", ar.MoveTaskToAnotherList())
					r.Patch("/move/heading", ar.MoveTaskToAnotherHeading())
					r.Patch("/complete", ar.CompleteTask())
					r.Patch("/archive", ar.ArchiveTask())
				})
			})

			r.Get("/tags", ar.GetTagsByUserID())
		})
	})

	return r
}
