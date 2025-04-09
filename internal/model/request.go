package model

import (
	"io"
	"net/http"
	"net/url"
)

type RequestData struct {
	Method      string              `json:"method"`
	Path        string              `json:"path"`
	QueryParams url.Values          `json:"queryParams"`
	Headers     map[string][]string `json:"headers"`
	Body        string              `json:"body"`
	RemoteAddr  string              `json:"remoteAddr"`
	Host        string              `json:"host"`
	Protocol    string              `json:"protocol"`
}

func ExtractRequestData(r *http.Request) (*RequestData, error) {
	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Create request data
	data := &RequestData{
		Method:      r.Method,
		Path:        r.URL.Path,
		QueryParams: r.URL.Query(),
		Headers:     r.Header,
		Body:        string(body),
		RemoteAddr:  r.RemoteAddr,
		Host:        r.Host,
		Protocol:    r.Proto,
	}

	return data, nil
}
