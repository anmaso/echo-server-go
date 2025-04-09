package handler

import (
	"encoding/json"
	"net/http"
	"path"

	"echo-server/internal/config"
	"echo-server/pkg/logger"
)

type ConfigurationHandler struct {
	configManager *config.ConfigManager
}

func NewConfigurationHandler(cm *config.ConfigManager) *ConfigurationHandler {
	return &ConfigurationHandler{
		configManager: cm,
	}
}

func (h *ConfigurationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r)
	case http.MethodPost:
		h.handlePost(w, r)
	case http.MethodPut:
		h.handlePut(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ConfigurationHandler) handleGet(w http.ResponseWriter, r *http.Request) {
	cfg := h.configManager.GetConfig()

	response := struct {
		ServerConfig *config.ServerConfig `json:"server"`
		Paths        []config.PathConfig  `json:"paths"`
	}{
		ServerConfig: cfg,
		Paths:        cfg.PathMatcher.GetAllConfigs(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode config response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *ConfigurationHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	var pathCfg config.PathConfig
	if err := json.NewDecoder(r.Body).Decode(&pathCfg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.configManager.UpdatePathConfig(pathCfg); err != nil {
		logger.Error("Failed to update path config: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *ConfigurationHandler) handlePut(w http.ResponseWriter, r *http.Request) {
	pattern := path.Base(r.URL.Path)
	if pattern == "" {
		http.Error(w, "Pattern not specified", http.StatusBadRequest)
		return
	}

	var pathCfg config.PathConfig
	if err := json.NewDecoder(r.Body).Decode(&pathCfg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pathCfg.Pattern = pattern
	if err := h.configManager.UpdatePathConfig(pathCfg); err != nil {
		logger.Error("Failed to update path config: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
