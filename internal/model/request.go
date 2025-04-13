package model

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RequestData struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	QueryParams url.Values        `json:"queryParams"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	RemoteAddr  string            `json:"remoteAddr"`
	Host        string            `json:"host"`
	Protocol    string            `json:"protocol"`
	Counter     CounterInfo       `json:"counter"`
}

type ResponseData struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Request    *RequestData      `json:"request"`
	Counter    CounterInfo       `json:"counter"`
}

type CounterInfo struct {
	Global uint64 `json:"global"`
	Path   uint64 `json:"path"`
}

func ExtractRequestData(r *http.Request) (*RequestData, error) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var headers = make(map[string]string)

	for key, value := range r.Header {
		headers[key] = strings.Join(value, ", ")
	}

	// Create request data
	data := &RequestData{
		Method:      r.Method,
		Path:        r.URL.Path,
		QueryParams: r.URL.Query(),
		Headers:     headers,
		Body:        string(body),
		RemoteAddr:  r.RemoteAddr,
		Host:        r.Host,
		Protocol:    r.Proto,
	}

	return data, nil
}
