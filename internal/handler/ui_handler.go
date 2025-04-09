package handler

import (
	"embed"
	"net/http"
	"strings"

	"echo-server/internal/config"
	"echo-server/pkg/logger"
)

//go:embed static/*
var staticFiles embed.FS

type UIHandler struct {
	configManager *config.ConfigManager
}

func NewUIHandler(cm *config.ConfigManager) http.Handler {
	return &UIHandler{
		configManager: cm,
	}
}

func (h *UIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Strip /ui prefix from path
	path := strings.TrimPrefix(r.URL.Path, "/ui")
	if path == "" || path == "/" {
		path = "/ui.html"
	}

	// Serve from embedded files
	content, err := staticFiles.ReadFile("static" + path)
	if err != nil {
		logger.Error("Failed to read static file: %v", err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Set content type based on file extension
	contentType := "text/plain"
	switch {
	case strings.HasSuffix(path, ".html"):
		contentType = "text/html"
	case strings.HasSuffix(path, ".css"):
		contentType = "text/css"
	case strings.HasSuffix(path, ".js"):
		contentType = "application/javascript"
	}
	w.Header().Set("Content-Type", contentType)
	w.Write(content)
}
