package v1

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/go-chi/render"

	"github.com/rshelekhov/reframed/internal/port"
	"github.com/rshelekhov/reframed/pkg/httpserver/middleware/jwtoken"
	mwlogger "github.com/rshelekhov/reframed/pkg/httpserver/middleware/logger"
	"github.com/rshelekhov/reframed/pkg/logger"
)

func NewRouter(
	log logger.Interface,
	jwt *jwtoken.TokenService,
	a port.AuthUsecase,
	l port.ListUsecase,
	h port.HeadingUsecase,
	t port.TaskUsecase,
	tag port.TagUsecase,
) *chi.Mux {
	r := chi.NewRouter()

	// Add request_id to each request, for tracing purposes
	r.Use(middleware.RequestID)

	// Logging of all requests
	r.Use(middleware.Logger)

	// By default, middleware.logger uses its own internal logger,
	// which should be overridden to use ours. Otherwise, problems
	// may arise - for example, with log collection. We can use
	// our own middleware to log requests:
	r.Use(mwlogger.New(log))

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

	NewAuthRoutes(r, log, jwt, a)
	NewListRoutes(r, log, jwt, l)
	NewHeadingRoutes(r, log, jwt, h)
	NewTaskRoutes(r, log, jwt, t)
	NewTagRoutes(r, log, jwt, tag)

	return r
}
