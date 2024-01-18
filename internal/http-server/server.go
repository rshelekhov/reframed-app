package http_server

import (
	"context"
	"errors"
	"github.com/go-chi/chi"
	"github.com/rshelekhov/reframed/config"
	"github.com/rshelekhov/reframed/internal/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	cfg    *config.Config
	log    logger.Interface
	Router *chi.Mux
}

func NewServer(cfg *config.Config, log logger.Interface, router *chi.Mux) *Server {
	srv := &Server{
		cfg:    cfg,
		log:    log,
		Router: router,
	}

	return srv
}

func (s *Server) Start() {
	srv := http.Server{
		Addr:         s.cfg.HTTPServer.Address,
		Handler:      s.Router,
		ReadTimeout:  s.cfg.HTTPServer.Timeout,
		WriteTimeout: s.cfg.HTTPServer.Timeout,
		IdleTimeout:  s.cfg.HTTPServer.IdleTimeout,
	}

	// TODO: move timeout to config
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdownComplete := handleShutdown(func() {
		if err := srv.Shutdown(ctx); err != nil {
			s.log.Error("server.Shutdown failed")
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
