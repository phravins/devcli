package taskrunner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectGoTasks(t *testing.T) {
	tempDir := t.TempDir()

	// Base tasks expected for any Go project
	baseTasks := []Task{
		{
			Name:        "Build Go Project",
			Type:        TaskBuild,
			Command:     "go build ./...",
			Description: "Build Go project",
			Icon:        "",
		},
		{
			Name:        "Run Tests",
			Type:        TaskTest,
			Command:     "go test ./...",
			Description: "Run all Go tests",
			Icon:        "",
		},
		{
			Name:        "Format Code (gofmt)",
			Type:        TaskFormat,
			Command:     "gofmt -w .",
			Description: "Format Go code",
			Icon:        "",
		},
		{
			Name:        "Run Go Vet",
			Type:        TaskLint,
			Command:     "go vet ./...",
			Description: "Check Go code for issues",
			Icon:        "",
		},
	}

	tests := []struct {
		name          string
		setup         func(string)
		expectedTasks []Task
	}{
		{
			name:          "Without main.go",
			setup:         func(dir string) {}, // No setup needed
			expectedTasks: baseTasks,
		},
		{
			name: "With main.go",
			setup: func(dir string) {
				file, err := os.Create(filepath.Join(dir, "main.go"))
				if err != nil {
					t.Fatalf("Failed to create main.go: %v", err)
				}
				file.Close()
			},
			expectedTasks: append(append([]Task(nil), baseTasks...), Task{
				Name:        "Run main.go",
				Type:        TaskRun,
				Command:     "go run main.go",
				Description: "Execute main.go",
				Icon:        "",
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a specific sub-directory for this test case to avoid interference
			testDir := filepath.Join(tempDir, tt.name)
			err := os.Mkdir(testDir, 0755)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			tt.setup(testDir)

			tasks := detectGoTasks(testDir)

			if len(tasks) != len(tt.expectedTasks) {
				t.Fatalf("Expected %d tasks, got %d", len(tt.expectedTasks), len(tasks))
			}

			for i, expectedTask := range tt.expectedTasks {
				if !reflect.DeepEqual(tasks[i], expectedTask) {
					t.Errorf("Mismatch at index %d.\nExpected: %+v\nGot: %+v", i, expectedTask, tasks[i])
				}
			}
		})
	}
}
