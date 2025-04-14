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

	"github.com/gorilla/mux"
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
	routes := mux.NewRouter()

	// Add rate limiting middleware that will be applied to all routes
	rateLimitMiddleware := middleware.RateLimit(configManager.GetConfig().PathMatcher)

	// Configuration endpoints
	configHandler := handler.NewConfigurationHandler(configManager)
	routes.PathPrefix("/config").Handler(middleware.RequestLogging(rateLimitMiddleware(configHandler)))

	// Counter endpoint with logging middleware
	routes.Handle("/counter", middleware.RequestLogging(rateLimitMiddleware(http.HandlerFunc(handler.CounterHandler))))

	// Main echo handler with logging middleware for all other paths
	uiHandler := handler.NewUIHandler(configManager)
	routes.Handle("/ui/", middleware.RequestLogging(rateLimitMiddleware(uiHandler)))
	routes.PathPrefix("/ui/").Handler(middleware.RequestLogging(rateLimitMiddleware(uiHandler)))
	routes.Handle("/ui", http.RedirectHandler("/ui/", http.StatusPermanentRedirect))
	routes.PathPrefix("/").Handler(middleware.RequestLogging(rateLimitMiddleware(handler.NewEchoHandler(configManager.GetConfig()))))

	return routes
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
