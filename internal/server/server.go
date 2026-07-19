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
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	rateLimit := middleware.RateLimit(configManager.GetConfig().PathMatcher)
	logging := middleware.RequestLoggingHandler()
	capture := middleware.RequestCaptureMiddleware()

	configHandler := handler.NewConfigurationHandler(configManager)
	routes.PathPrefix("/config").Handler(logging(capture(configHandler)))

	routes.Handle("/counter", logging(capture(handler.NewCounterHandler())))

	// Request log endpoints (no capture — we don't log requests about the log).
	logHandler := handler.NewLogHandler()
	routes.HandleFunc("/ui/requests/stream", logHandler.HandleStream).Methods(http.MethodGet)
	routes.HandleFunc("/ui/requests/config", logHandler.HandleConfig).Methods(http.MethodGet, http.MethodPut)
	routes.HandleFunc("/ui/requests", logHandler.HandleList).Methods(http.MethodGet)
	routes.HandleFunc("/ui/requests", logHandler.HandleClear).Methods(http.MethodDelete)

	uiHandler := handler.NewUIHandler(configManager)
	routes.Handle("/ui/", logging(uiHandler))
	routes.PathPrefix("/ui/").Handler(logging(uiHandler))
	routes.Handle("/ui", http.RedirectHandler("/ui/ui.html", http.StatusPermanentRedirect))

	routes.Handle("/metrics", promhttp.Handler())

	routes.PathPrefix("/").Handler(logging(capture(rateLimit(middleware.MetricsMiddleware(handler.NewEchoHandler(configManager.GetConfig()))))))

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
