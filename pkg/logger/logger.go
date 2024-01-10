package logger

import (
	"github.com/go-chi/chi/middleware"
	"log/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type Interface interface {
	With(attrs ...interface{}) *Logger
	Debug(msg string, attrs ...interface{})
	Info(msg string, attrs ...interface{})
	Warn(msg string, attrs ...interface{})
	Error(msg string, attrs ...interface{})
}

type Logger struct {
	logger *slog.Logger
}

func SetupLogger(env string) *Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return &Logger{logger: log}
}

// Err ...
func Err(err error) slog.Attr {
	return slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	}
}

func LogWithRequest(log Interface, op string, r *http.Request) Interface {
	log.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)
	return log
}

func (l *Logger) With(attrs ...interface{}) *Logger {
	l.logger.With(attrs...)
	return l
}

func (l *Logger) Debug(msg string, attrs ...interface{}) {
	l.logger.Debug(msg, attrs...)
}

func (l *Logger) Info(msg string, attrs ...interface{}) {
	l.logger.Info(msg, attrs...)
}

func (l *Logger) Warn(msg string, attrs ...interface{}) {
	l.logger.Warn(msg, attrs...)
}

func (l *Logger) Error(msg string, attrs ...interface{}) {
	l.logger.Error(msg, attrs...)
}
