package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"echo-server/internal/config"
	"echo-server/internal/handler"
	"echo-server/internal/middleware"
	"echo-server/internal/model" // Added for HistoryStorage
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
	// Get the loaded server configuration
	serverCfg := configManager.GetConfig()

	// Initialize HistoryStorage with DefaultMaxSize from the configuration.
	// Defaults for HistoryConfig (Enabled: false, DefaultMaxSize: 100) are applied in the loader.
	historyStorage := model.NewHistoryStorage(serverCfg.History.DefaultMaxSize)
	logger.Info("HistoryStorage initialized with MaxSize: %d. Recording initially: %v",
		serverCfg.History.DefaultMaxSize, historyStorage.IsRecordingActive())
	// serverCfg.History.Enabled is not used to auto-start recording here,
	// but signifies that the feature is available. API/UI calls will start/stop recording.

	return &Server{
		configManager: configManager,
		// Pass historyStorage to setupRoutes
		handler:       setupRoutes(configManager, historyStorage),
	}
}

func setupRoutes(configManager *config.ConfigManager, historyStorage *model.HistoryStorage) http.Handler {
	routes := mux.NewRouter()

	rateLimit := middleware.RateLimit(configManager.GetConfig().PathMatcher)
	// Standard logging for most routes
	logging := middleware.RequestLoggingHandler()
	// Logging and History handler for the main echo path
	loggingAndHistory := middleware.LoggingAndHistoryHandler(historyStorage)

	// Configuration Handler
	configHandler := handler.NewConfigurationHandler(configManager)
	routes.PathPrefix("/config").Handler(logging(configHandler))

	// Counter Handler
	routes.Handle("/counter", logging(handler.NewCounterHandler()))

	// UI Handler
	uiHandler := handler.NewUIHandler(configManager)
	routes.Handle("/ui/", logging(uiHandler))          // Specific path for /ui/
	routes.PathPrefix("/ui/").Handler(logging(uiHandler)) // Catch all under /ui/
	routes.Handle("/ui", http.RedirectHandler("/ui/ui.html", http.StatusPermanentRedirect))


	// Metrics Handler (usually doesn't need rich logging)
	routes.Handle("/metrics", promhttp.Handler())

	// History Handler
	// All /history routes are managed by HistoryHandler's ServeHTTP
	// Using standard logging for history management endpoints to avoid recursive history entries.
	historyHandler := handler.NewHistoryHandler(historyStorage, configManager)
	routes.PathPrefix("/history").Handler(logging(historyHandler))


	// Main Echo Handler - now uses LoggingAndHistoryHandler
	// IMPORTANT: The order of middleware: LoggingAndHistory -> RateLimit -> Metrics -> Actual Handler
	echoHandler := handler.NewEchoHandler(configManager.GetConfig())
	mainHandlerChain := loggingAndHistory(rateLimit(middleware.MetricsMiddleware(echoHandler)))
	routes.PathPrefix("/").Handler(mainHandlerChain)

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
