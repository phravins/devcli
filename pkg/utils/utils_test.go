package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindExecutable(t *testing.T) {
	// Create a temporary directory for fallback testing
	tempDir, err := os.MkdirTemp("", "find_exec_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy executable file in the temp directory
	dummyExec := filepath.Join(tempDir, "dummy_exec_test_xyz_123")
	f, err := os.Create(dummyExec)
	if err != nil {
		t.Fatalf("failed to create dummy exec: %v", err)
	}
	f.Close()
	// On Windows, the actual executable name often needs a .exe or similar, but
	// FindExecutable uses filepath.Glob which just looks for matching files.
	// We'll test the glob match.

	t.Run("Command in PATH", func(t *testing.T) {
		// "go" or "ls" (unix) or "cmd" (windows) should be in PATH.
		// Since we run go test, "go" is definitely in PATH.
		found := FindExecutable("go", nil)
		if found == "" {
			t.Error("expected 'go' to be found in PATH, got empty string")
		}
	})

	t.Run("Command not in PATH, found via fallback", func(t *testing.T) {
		// We use a command name that is highly unlikely to be in PATH
		cmdName := "nonexistent_command_12345"

		// Create a fallback glob that points to our dummy executable
		fallbackGlobs := []string{
			filepath.Join(tempDir, "dummy_exec_test_*"),
		}

		found := FindExecutable(cmdName, fallbackGlobs)
		if found != dummyExec {
			t.Errorf("expected to find %s via fallback, got %s", dummyExec, found)
		}
	})

	t.Run("Command not found anywhere", func(t *testing.T) {
		cmdName := "nonexistent_command_54321"
		fallbackGlobs := []string{
			filepath.Join(tempDir, "does_not_exist_*"),
		}

		found := FindExecutable(cmdName, fallbackGlobs)
		if found != "" {
			t.Errorf("expected not to find anything, got %s", found)
		}
	})
}
