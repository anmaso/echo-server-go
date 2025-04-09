package server

import (
	"context"
	"fmt"
	"net/http"

	"echo-server/internal/config"
	"echo-server/internal/handler"
	"echo-server/internal/middleware"
	"echo-server/pkg/logger"
)

type Server struct {
	server  *http.Server
	config  *config.ServerConfig
	handler *handler.EchoHandler
}

func New(cfg *config.ServerConfig) *Server {
	if cfg == nil {
		cfg = config.Default()
	}

	return &Server{
		config:  cfg,
		handler: handler.New(cfg),
	}
}

func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Counter endpoint with logging middleware
	mux.Handle("/counter", middleware.RequestLogging(http.HandlerFunc(handler.CounterHandler)))

	// Main echo handler with logging middleware for all other paths
	mux.Handle("/", middleware.RequestLogging(s.handler))

	return mux
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	handler := s.setupRoutes()

	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  s.config.ReadTimeout.Duration,
		WriteTimeout: s.config.WriteTimeout.Duration,
	}

	logger.Info("Starting server on %s", addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		logger.Info("Shutting down server...")
		return s.server.Shutdown(ctx)
	}
	return nil
}
