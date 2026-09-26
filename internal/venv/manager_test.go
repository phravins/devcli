package venv

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckPrerequisites_Success(t *testing.T) {
	candidates := []string{"python", "python3", "py"}

	for _, candidate := range candidates {
		t.Run(candidate, func(t *testing.T) {
			tempDir := t.TempDir()

			dummyName := candidate
			if runtime.GOOS == "windows" {
				dummyName += ".exe"
			}

			dummyPath := filepath.Join(tempDir, dummyName)
			err := os.WriteFile(dummyPath, []byte("#!/bin/sh\nexit 0"), 0755)
			if err != nil {
				t.Fatalf("failed to create dummy %s: %v", candidate, err)
			}

			t.Setenv("PATH", tempDir)

			m := NewManager("")
			err = m.CheckPrerequisites()
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if m.PythonPath != dummyPath {
				t.Errorf("expected PythonPath %q, got %q", dummyPath, m.PythonPath)
			}
		})
	}
}

func TestCheckPrerequisites_Failure(t *testing.T) {
	tempDir := t.TempDir()

	t.Setenv("PATH", tempDir)

	m := NewManager("")
	err := m.CheckPrerequisites()
	if err == nil {
		t.Errorf("expected error, got nil")
	}

	expectedErr := "python is not installed or not in PATH (tried python, python3, py)"
	if err.Error() != expectedErr {
		t.Errorf("expected error %q, got %q", expectedErr, err.Error())
	}
}
