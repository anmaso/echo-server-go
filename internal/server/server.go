package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"echo-server/internal/config"
	"echo-server/internal/handler"
	"echo-server/internal/middleware"
	"echo-server/pkg/logger"
)

type Server struct {
	configManager *config.ConfigManager
	server        *http.Server
	mu            sync.RWMutex
	handler       *handler.EchoHandler
}

func New(configManager *config.ConfigManager) *Server {
	return &Server{
		configManager: configManager,
		handler:       handler.NewEchoHandler(configManager.GetConfig()),
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
	cfg := s.configManager.GetConfig()
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	handler := s.setupRoutes()

	s.mu.Lock()
	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout.Duration,
		WriteTimeout: cfg.WriteTimeout.Duration,
	}
	s.mu.Unlock()

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
