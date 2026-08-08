package unit

import (
	"os"
	"path/filepath"
	"testing"

	"omnia/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.Workers < 1 || cfg.Workers > 6 {
		t.Errorf("expected default workers between 1 and 6, got %d", cfg.Workers)
	}
	if cfg.Compression != "balanced" {
		t.Errorf("expected default compression 'balanced', got %s", cfg.Compression)
	}
	if cfg.OutputDirectory != "." {
		t.Errorf("expected default output_directory '.', got %s", cfg.OutputDirectory)
	}
}

func TestLoadConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	cfgPath := filepath.Join(tempDir, "config.yaml")

	content := []byte(`
workers: 4
compression: high
output_directory: /tmp/omnia_out
keep_original: false
overwrite: true
log_level: debug
`)
	if err := os.WriteFile(cfgPath, content, 0644); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		t.Fatalf("LoadConfig returned unexpected error: %v", err)
	}

	if cfg.Workers != 4 {
		t.Errorf("expected Workers 4, got %d", cfg.Workers)
	}
	if cfg.Compression != "high" {
		t.Errorf("expected Compression 'high', got %s", cfg.Compression)
	}
	if cfg.OutputDirectory != "/tmp/omnia_out" {
		t.Errorf("expected OutputDirectory '/tmp/omnia_out', got %s", cfg.OutputDirectory)
	}
	if cfg.KeepOriginal != false {
		t.Errorf("expected KeepOriginal false, got %t", cfg.KeepOriginal)
	}
	if cfg.Overwrite != true {
		t.Errorf("expected Overwrite true, got %t", cfg.Overwrite)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug', got %s", cfg.LogLevel)
	}
}
