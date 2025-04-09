package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoader(t *testing.T) {
	// Create temporary test directory
	tmpDir, err := os.MkdirTemp("", "config-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test server config
	serverConfig := `{
        "host": "localhost",
        "port": 8080,
        "readTimeout": "30s",
        "writeTimeout": "30s",
        "defaultResponse": {
            "statusCode": 200,
            "headers": {
                "Content-Type": "application/json"
            }
        }
    }`

	if err := os.WriteFile(filepath.Join(tmpDir, "server.json"), []byte(serverConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test path config
	pathConfig := `{
        "pattern": "^/test/.*",
        "methods": ["GET", "POST"],
        "response": {
            "statusCode": 200,
            "headers": {
                "Content-Type": "application/json"
            },
            "body": "{\"status\":\"ok\"}",
            "delay": "100ms"
        },
        "counterEnabled": true
    }`

	pathsDir := filepath.Join(tmpDir, "paths")
	if err := os.Mkdir(pathsDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(pathsDir, "test.json"), []byte(pathConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Test loading configurations
	loader := NewLoader()

	t.Run("LoadServerConfig", func(t *testing.T) {
		err := loader.LoadServerConfig(filepath.Join(tmpDir, "server.json"))
		if err != nil {
			t.Errorf("LoadServerConfig failed: %v", err)
		}

		cfg := loader.GetConfig()
		if cfg.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", cfg.Port)
		}
	})

	t.Run("LoadPathConfigs", func(t *testing.T) {
		err := loader.LoadPathConfigs(pathsDir)
		if err != nil {
			t.Errorf("LoadPathConfigs failed: %v", err)
		}

		cfg := loader.GetConfig()
		match, _ := cfg.PathMatcher.Match("/test/123", "GET")
		if match == nil {
			t.Error("Expected to find matching path config")
		}
	})
}
