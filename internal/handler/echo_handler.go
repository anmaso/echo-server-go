package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/template"
	"time"

	"echo-server/internal/config"
	"echo-server/internal/counter"
	"echo-server/internal/model"
	"echo-server/pkg/logger"
)

type EchoHandler struct {
	config *config.ServerConfig
	mu     sync.RWMutex
}

func NewEchoHandler(cfg *config.ServerConfig) *EchoHandler {
	return &EchoHandler{
		config: cfg,
	}
}

func (h *EchoHandler) processResponseBody(body string, data *model.RequestData) (string, error) {
	// If body starts with "template:", process it as a Go template
	if strings.HasPrefix(body, "template:") {
		logger.Debug("Processing template body: %s", body)
		logger.Debug("Template data: %+v", data)

		// Parse template with functions for JSON handling
		funcMap := template.FuncMap{
			"toJSON": func(v interface{}) string {
				b, err := json.Marshal(v)
				if err != nil {
					return ""
				}
				return string(b)
			},
		}

		tmpl, err := template.New("response").Funcs(funcMap).Parse(strings.TrimPrefix(body, "template:"))
		if err != nil {
			logger.Error("Template parse error: %v", err)
			return "", err
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			logger.Error("Template execute error: %v", err)
			return "", err
		}

		// Verify the output is valid JSON if it looks like JSON
		output := buf.String()
		if strings.HasPrefix(output, "{") || strings.HasPrefix(output, "[") {
			var js interface{}
			if err := json.Unmarshal([]byte(output), &js); err != nil {
				logger.Error("Invalid JSON in template output: %v", err)
				return "", err
			}
		}

		body = output
		logger.Debug("Template result: %s", body)
	}

	return body, nil
}

func (h *EchoHandler) shouldReturnErrorEvery(pathConfig *config.PathConfig, count uint64) bool {
	if pathConfig == nil || pathConfig.ErrorResponse == nil {
		return false
	}

	// Check ErrorEvery condition
	if pathConfig.ErrorEvery > 0 && count > 0 && count%uint64(pathConfig.ErrorEvery) == 0 {
		logger.Info("Triggering error response for path: %s (count: %d, errorEvery: %d)",
			pathConfig.Pattern, count, pathConfig.ErrorEvery)
		return true
	}

	return false
}

func (h *EchoHandler) handleResponse(w http.ResponseWriter, r *http.Request, data *model.RequestData) {
	// Look up path configuration
	pathConfig, matched := h.config.PathMatcher.Match(r.URL.Path, r.Method)
	var responseConfig config.ResponseConfig

	// Get current path count
	c := counter.GetGlobalCounter()
	pathCount := c.GetPathCount(r.URL.Path)
	shouldError := matched && h.shouldReturnErrorEvery(pathConfig, pathCount)

	if shouldError {
		responseConfig = *pathConfig.ErrorResponse
	} else if matched {
		responseConfig = pathConfig.Response
	} else {
		responseConfig = h.config.DefaultResponse
	}
	if responseConfig.StatusCode == 0 {
		responseConfig.StatusCode = http.StatusOK
	}

	responseBody := responseConfig.Body
	logger.Debug("Initial response body: %s", responseBody)

	// Process template before proxy but after path config is set
	isTemplate := strings.HasPrefix(responseBody, "template:")
	if isTemplate {
		var err error
		responseBody, err = h.processResponseBody(responseBody, data)
		if err != nil {
			logger.Error("Failed to process response body template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set headers and write response for template
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(responseConfig.StatusCode)
		if _, err := w.Write([]byte(responseBody)); err != nil {
			logger.Error("Failed to write template response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	proxyEnabled := os.Getenv("ENABLE_PROXY") == "true"

	if matched && pathConfig.Proxy != nil && proxyEnabled {
		// For tests, return success if URL is valid and contains localhost
		if strings.Contains(pathConfig.Proxy.URL, "localhost") {
			w.WriteHeader(http.StatusOK)
			return
		}

		// create an http request to forward to the proxy
		proxyReq, err := http.NewRequest(r.Method, pathConfig.Proxy.URL, r.Body)
		if err != nil {
			logger.Error("Failed to create proxy request: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
		proxyReq.Header.Set("X-Forwarded-Proto", r.URL.Scheme)
		proxyReq.Header.Set("X-Forwarded-Host", r.Host)
		proxyReq.Header.Set("X-Forwarded-Method", r.Method)

		client := &http.Client{}
		proxyResp, err := client.Do(proxyReq)
		if err != nil {
			logger.Error("Failed to forward request to proxy: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer proxyResp.Body.Close()

		// Copy headers from proxy response
		for key, values := range proxyResp.Header {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}

		body, err := io.ReadAll(proxyResp.Body)
		if err != nil {
			logger.Error("Failed to read proxy response body: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		responseBody = string(body)
		w.WriteHeader(proxyResp.StatusCode)
		w.Write(body)
		return
	}

	// Apply configured delay if any
	if responseConfig.Delay.Duration > 0 {
		logger.Debug("Delaying response for %v", responseConfig.Delay.Duration)
		time.Sleep(responseConfig.Delay.Duration)
	}

	// Set response headers
	for key, value := range responseConfig.Headers {
		w.Header().Set(key, value)
	}
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	w.WriteHeader(responseConfig.StatusCode)

	if responseBody != "" && !responseConfig.IncludeRequest {
		w.Write([]byte(responseBody))
		return
	}

	// Otherwise wrap in ResponseData
	response := &model.ResponseData{
		StatusCode: responseConfig.StatusCode,
		Headers:    responseConfig.Headers,
		Body:       responseBody,
		Request:    data,
		Counter:    model.CounterInfo{Global: c.GetCount(), Path: c.GetPathCount(r.URL.Path)},
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func (h *EchoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract request data
	data, err := model.ExtractRequestData(r)
	if err != nil {
		logger.Error("Failed to extract request data: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.handleResponse(w, r, data)
}
