package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"echo-server/internal/config"
	"echo-server/pkg/logger"

	"github.com/samber/lo"
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
	segments := strings.Split(r.URL.Path+"/", "/")

	switch {
	case r.Method == http.MethodGet:
		h.handleGet(w, r, segments[2])
	case r.Method == http.MethodPost:
		h.handlePost(w, r)
	case r.Method == http.MethodDelete:
		h.handleDelete(w, r, segments[2])
	case r.Method == http.MethodPut:
		h.handlePut(w, r, segments[2])
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ConfigurationHandler) handleGet(w http.ResponseWriter, r *http.Request, name string) {
	cfg := h.configManager.GetConfig()
	configs := lo.Filter(cfg.PathMatcher.GetAllConfigs(), func(item config.PathConfig, _ int) bool {
		return name == "" || item.Name == name
	})

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(configs); err != nil {
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

func (h *ConfigurationHandler) handlePut(w http.ResponseWriter, r *http.Request, name string) {
	var pathCfg config.PathConfig
	if err := json.NewDecoder(r.Body).Decode(&pathCfg); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pathCfg.Pattern = name
	if err := h.configManager.UpdatePathConfig(pathCfg); err != nil {
		logger.Error("Failed to update path config: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ConfigurationHandler) handleDelete(w http.ResponseWriter, r *http.Request, name string) {
	if deleted := h.configManager.GetConfig().PathMatcher.DeleteByName(name); !deleted {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
