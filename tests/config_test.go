package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/reversTeam/go-ms/core"
)

func TestNewConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("/tmp/", "go-ms-test-dir")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile := filepath.Join(tmpDir, "config.yaml")
	configContent := []byte(`
name: test-service
grpc:
  host: localhost
  port: 8080
http:
  host: localhost
  port: 8081
  read-timeout: 30
  write-timeout: 30
exporter:
  host: localhost
  port: 9090
  path: /metrics
  interval: 15
jaeger:
  name: jaeger-service
  unsecure: false
  mode: mock
  host: localhost
  port: 5775
  sampler:
    type: const
    param: 1
  reporter:
    log-spans: true
services:
  user:
    http: true
    config:
      baseURL: "http://localhost:8082"
middlewares:
  - auth
  - logging
`)
	err = os.WriteFile(tmpFile, configContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Execute the function under test.
	config, err := core.NewConfig(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Assertions
	if config.Name != "test-service" {
		t.Errorf("Expected name to be 'test-service', got '%s'", config.Name)
	}
	if config.Grpc.Host != "localhost" || config.Grpc.Port != 8080 {
		t.Errorf("GRPC config incorrect, got: %v", config.Grpc)
	}
	if config.Http.Host != "localhost" || config.Http.Port != 8081 || config.Http.ReadTimeout != 60 || config.Http.WriteTimeout != 60 {
		t.Errorf("HTTP config incorrect, got: %v", config.Http)
	}
	if config.Exporter.Host != "localhost" || config.Exporter.Port != 9090 || config.Exporter.Path != "/metrics" || config.Exporter.Interval != 15 {
		t.Errorf("Exporter config incorrect, got: %v", config.Exporter)
	}
	// if config.Jaeger.Name != "jaeger-service" || config.Jaeger.Mode != "mock" || config.Jaeger.Host != "localhost" || config.Jaeger.Port != 5775 || !config.Jaeger.Reporter.LogSpans {
	// 	t.Errorf("Jaeger config incorrect, got: %v", config.Jaeger)
	// }
	if len(config.Services) != 1 || !config.Services["user"].Http {
		t.Errorf("Services config incorrect, got: %v", config.Services)
	}
	// if len(config.Middleware) != 2 || config.Middleware[0] != "auth" || config.Middleware[1] != "logging" {
	// 	t.Errorf("Middleware config incorrect, got: %v", config.Middleware)
	// }
}
