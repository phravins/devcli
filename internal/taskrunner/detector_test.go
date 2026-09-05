package taskrunner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGoTasks(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "taskrunner_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test case 1: Directory without main.go
	t.Run("Without main.go", func(t *testing.T) {
		tasks := detectGoTasks(tmpDir)
		if len(tasks) != 4 {
			t.Errorf("expected 4 tasks without main.go, got %d", len(tasks))
		}

		// Verify expected tasks are present
		expectedTasks := []string{"Build Go Project", "Run Tests", "Format Code (gofmt)", "Run Go Vet"}
		for i, expected := range expectedTasks {
			if i < len(tasks) && tasks[i].Name != expected {
				t.Errorf("expected task at index %d to be %q, got %q", i, expected, tasks[i].Name)
			}
		}
	})

	// Test case 2: Directory with main.go
	t.Run("With main.go", func(t *testing.T) {
		mainGoPath := filepath.Join(tmpDir, "main.go")
		err := os.WriteFile(mainGoPath, []byte("package main\n\nfunc main() {}\n"), 0644)
		if err != nil {
			t.Fatalf("failed to create main.go: %v", err)
		}

		tasks := detectGoTasks(tmpDir)
		if len(tasks) != 5 {
			t.Errorf("expected 5 tasks with main.go, got %d", len(tasks))
		}

		// Verify the 5th task is "Run main.go"
		if len(tasks) == 5 && tasks[4].Name != "Run main.go" {
			t.Errorf("expected 5th task to be %q, got %q", "Run main.go", tasks[4].Name)
		}
	})
}
