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
	handler       http.Handler
}

func New(configManager *config.ConfigManager) *Server {
	return &Server{
		configManager: configManager,
		handler:       setupRoutes(configManager),
	}
}

func setupRoutes(configManager *config.ConfigManager) http.Handler {
	mux := http.NewServeMux()

	// Configuration endpoints
	configHandler := handler.NewConfigurationHandler(configManager)
	mux.Handle("/config", middleware.RequestLogging(configHandler))

	// Counter endpoint with logging middleware
	mux.Handle("/counter", middleware.RequestLogging(http.HandlerFunc(handler.CounterHandler)))

	// Main echo handler with logging middleware for all other paths
	mux.Handle("/", middleware.RequestLogging(handler.NewEchoHandler(configManager.GetConfig())))
	uiHandler := handler.NewUIHandler(configManager)
	mux.Handle("/ui/", middleware.RequestLogging(uiHandler))
	mux.Handle("/ui", http.RedirectHandler("/ui/", http.StatusPermanentRedirect))

	return mux
}

func (s *Server) Start() error {
	cfg := s.configManager.GetConfig()
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	s.mu.Lock()
	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.handler,
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
