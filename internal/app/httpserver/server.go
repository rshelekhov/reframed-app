package httpserver

import (
	"context"
	"errors"
	"github.com/rshelekhov/reframed/internal/config"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rshelekhov/reframed/internal/lib/middleware/jwtoken"

	"github.com/go-chi/chi/v5"

	"github.com/rshelekhov/reframed/internal/lib/logger"
)

type Server struct {
	cfg       *config.ServerSettings
	log       logger.Interface
	tokenAuth *jwtoken.TokenService
	router    *chi.Mux
}

func NewServer(
	cfg *config.ServerSettings,
	log logger.Interface,
	tokenAuth *jwtoken.TokenService,
	router *chi.Mux,
) *Server {
	srv := &Server{
		cfg:       cfg,
		log:       log,
		tokenAuth: tokenAuth,
		router:    router,
	}

	return srv
}

func (s *Server) Start() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	srv := http.Server{
		Addr:         s.cfg.HTTPServer.Address,
		Handler:      s.router,
		ReadTimeout:  s.cfg.HTTPServer.Timeout,
		WriteTimeout: s.cfg.HTTPServer.Timeout,
		IdleTimeout:  s.cfg.HTTPServer.IdleTimeout,
	}

	shutdownComplete := handleShutdown(func() {
		if err := srv.Shutdown(ctx); err != nil {
			s.log.Error("httpserver.Shutdown failed")
		}
	})

	if err := srv.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
		<-shutdownComplete
	} else {
		s.log.Error("http.ListenAndServe failed")
	}

	s.log.Info("shutdown gracefully")
}

func handleShutdown(onShutdownSignal func()) <-chan struct{} {
	shutdown := make(chan struct{})

	go func() {
		shutdownSignal := make(chan os.Signal, 1)
		signal.Notify(shutdownSignal, os.Interrupt, syscall.SIGTERM)

		<-shutdownSignal

		onShutdownSignal()
		close(shutdown)
	}()

	return shutdown
}
