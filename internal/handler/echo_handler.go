package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func (h *EchoHandler) processResponseBody(body string, data *model.RequestData) (interface{}, error) {
	// If body starts with "template:", process it as a Go template
	if strings.HasPrefix(body, "template:") {
		logger.Debug("Processing response body as template")
		tmpl, err := template.New("response").Parse(strings.TrimPrefix(body, "template:"))
		if err != nil {
			return nil, err
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return nil, err
		}
		body = buf.String()
	}

	// Try to parse as JSON
	var result interface{}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		// If not valid JSON, return as string
		logger.Debug("Response body is not valid JSON, returning as string")
		return body, nil
	}
	logger.Debug("Parsed response body as JSON: %v", result)
	return result, nil
}

func (h *EchoHandler) shouldReturnError(pathConfig *config.PathConfig, count uint64) bool {
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
	// Get counter instance
	c := counter.GetGlobalCounter()

	// Look up path configuration
	pathConfig, matched := h.config.PathMatcher.Match(r.URL.Path, r.Method)
	var responseConfig config.ResponseConfig

	if pathConfig.Proxy != nil {
		// create an http requet to forward to the proxy
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

		/*
			for key, value := range proxyResp.Header {
				w.Header()[key] = value
			}
			w.WriteHeader(proxyResp.StatusCode)
		*/
		body, err := io.ReadAll(proxyResp.Body)
		fmt.Println("%v", body)
		if err != nil {
			logger.Error("Failed to read proxy response body: %v", err)
		} else {
			data.Body = string(body)
			logger.Debug("Received proxy response: %s", data.Body)

		}
	}

	// Get current path count
	pathCount := c.GetPathCount(r.URL.Path)

	shouldError := matched && h.shouldReturnError(pathConfig, pathCount)

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

	// Process response body
	var responseBody interface{}
	if responseConfig.Body != "" {
		var err error
		responseBody, err = h.processResponseBody(responseConfig.Body, data)
		logger.Debug("Processed response body: %v", responseBody)
		if err != nil {
			logger.Error("Failed to process response body: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		// Default to echo request data if no body specified
		responseBody = data
	}

	response := struct {
		Request  *model.RequestData `json:"request,omitempty"`
		Response interface{}        `json:"response"`
		Status   int                `json:"status"`
	}{
		Response: responseBody,
		Status:   responseConfig.StatusCode,
	}

	// Include request data only if specified
	if responseConfig.IncludeRequest {
		response.Request = data
	}

	// Set status code and write response
	w.WriteHeader(responseConfig.StatusCode)
	if responseBody != "" {
		if str, ok := responseBody.(string); ok {
			w.Write([]byte(str))
		} else {
			if err := json.NewEncoder(w).Encode(responseBody); err != nil {
				logger.Error("Failed to encode response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		}
	} else {
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("Failed to encode response: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
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
