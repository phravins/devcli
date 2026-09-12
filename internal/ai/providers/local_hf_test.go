package providers

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/phravins/devcli/internal/config"
)

func TestLocalHFProvider_Metadata(t *testing.T) {
	p := &LocalHFProvider{}
	if got := p.Name(); got != "Local (Python/HF)" {
		t.Errorf("Expected Name() = 'Local (Python/HF)', got '%s'", got)
	}
	if got := p.Model(); got != "local-model" {
		t.Errorf("Expected Model() = 'local-model', got '%s'", got)
	}
	if got := p.IsLocal(); !got {
		t.Errorf("Expected IsLocal() = true, got %v", got)
	}
}

func TestLocalHFProvider_Configure_Success(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("PATH", tempDir)

	execName := "python"
	if runtime.GOOS == "windows" {
		execName = "python.exe"
	}
	filePath := filepath.Join(tempDir, execName)
	if err := os.WriteFile(filePath, []byte(""), 0755); err != nil {
		t.Fatalf("Failed to create dummy python binary: %v", err)
	}

	p := &LocalHFProvider{}
	cfg := &config.Config{}
	if err := p.Configure(cfg); err != nil {
		t.Errorf("Expected Configure to succeed, got: %v", err)
	}
}

func TestLocalHFProvider_Configure_PythonNotFound(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("PATH", tempDir)

	p := &LocalHFProvider{}
	cfg := &config.Config{}
	err := p.Configure(cfg)
	if err == nil {
		t.Fatal("Expected error when python is not on PATH, got nil")
	}

	expectedMsg := "python is not installed. Required for local HF models"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}
