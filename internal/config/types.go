package config

import (
	"encoding/json"
	"fmt"
	"time"
)

// Duration is a wrapper for time.Duration that implements JSON marshaling/unmarshaling
type Duration struct {
	time.Duration
}

// UnmarshalJSON implements custom JSON unmarshaling for Duration
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
	case float64:
		d.Duration = time.Duration(value)
	default:
		return fmt.Errorf("invalid duration")
	}
	return nil
}

// MarshalJSON implements custom JSON marshaling for Duration
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// ServerConfig holds the main server configuration
type ServerConfig struct {
	Host            string         `json:"host"`
	Port            int            `json:"port"`
	ReadTimeout     Duration       `json:"readTimeout"`
	WriteTimeout    Duration       `json:"writeTimeout"`
	DefaultResponse ResponseConfig `json:"defaultResponse"`
	PathMatcher     PathMatcher    `json:"-"`
	Paths           []PathConfig   `json:"paths"`
}

// Response represents the configuration for an HTTP response
type Response struct {
	StatusCode int      `json:"statusCode"`
	Headers    Headers  `json:"headers"`
	Body       string   `json:"body"`
	Delay      Duration `json:"delay"`
}

// Headers represents HTTP headers as key-value pairs
type Headers map[string]string
