package server

import (
	"context"
	"errors"
	"github.com/rshelekhov/reframed/src/logger"
	"github.com/rshelekhov/reframed/src/models"
	"github.com/rshelekhov/reframed/src/server/middleware/jwtoken/service"
	"github.com/rshelekhov/reframed/src/web/api"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	cfg       *models.ServerSettings
	log       logger.Interface
	tokenAuth *service.JWTokenService
	user      *api.UserHandler
	list      *api.ListHandler
	task      *api.TaskHandler
	heading   *api.HeadingHandler
	tag       *api.TagHandler
}

func NewServer(
	cfg *models.ServerSettings,
	log logger.Interface,
	tokenAuth *service.JWTokenService,
	user *api.UserHandler,
	list *api.ListHandler,
	task *api.TaskHandler,
	heading *api.HeadingHandler,
	tag *api.TagHandler,
) *Server {
	srv := &Server{
		cfg:       cfg,
		log:       log,
		tokenAuth: tokenAuth,
		list:      list,
		user:      user,
		task:      task,
		heading:   heading,
		tag:       tag,
	}

	return srv
}

func (s *Server) Start() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	routes := s.initRoutes(s.tokenAuth)

	srv := http.Server{
		Addr:         s.cfg.HTTPServer.Address,
		Handler:      routes,
		ReadTimeout:  s.cfg.HTTPServer.Timeout,
		WriteTimeout: s.cfg.HTTPServer.Timeout,
		IdleTimeout:  s.cfg.HTTPServer.IdleTimeout,
	}

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
