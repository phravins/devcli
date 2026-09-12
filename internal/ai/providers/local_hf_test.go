package providers

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/phravins/devcli/internal/config"
)

func TestLocalHFProvider_Getters(t *testing.T) {
	p := &LocalHFProvider{}

	if name := p.Name(); name != "Local (Python/HF)" {
		t.Errorf("Expected Name() to be 'Local (Python/HF)', got %q", name)
	}

	if model := p.Model(); model != "local-model" {
		t.Errorf("Expected Model() to be 'local-model', got %q", model)
	}

	if !p.IsLocal() {
		t.Errorf("Expected IsLocal() to be true, got false")
	}
}

func TestLocalHFProvider_Configure_PythonNotFound(t *testing.T) {
	// Set PATH to an empty temporary directory so python is not found
	tempDir := t.TempDir()
	t.Setenv("PATH", tempDir)

	p := &LocalHFProvider{}
	cfg := &config.Config{}

	err := p.Configure(cfg)
	if err == nil {
		t.Fatalf("Expected Configure() to fail when python is not in PATH, but got nil")
	}

	expectedErr := "python is not installed. Required for local HF models"
	if err.Error() != expectedErr {
		t.Errorf("Expected error message %q, got %q", expectedErr, err.Error())
	}
}

func TestLocalHFProvider_Configure_PythonFound(t *testing.T) {
	// Create a temporary directory containing a mock python executable
	tempDir := t.TempDir()
	pythonName := "python"
	if runtime.GOOS == "windows" {
		pythonName = "python.exe"
	}

	pythonPath := filepath.Join(tempDir, pythonName)
	if err := os.WriteFile(pythonPath, []byte("#!/bin/sh\necho 1"), 0755); err != nil {
		t.Fatalf("Failed to create dummy python executable: %v", err)
	}

	t.Setenv("PATH", tempDir)

	p := &LocalHFProvider{}
	cfg := &config.Config{}

	err := p.Configure(cfg)
	if err != nil {
		t.Fatalf("Expected Configure() to succeed when python is in PATH, got error: %v", err)
	}
}
