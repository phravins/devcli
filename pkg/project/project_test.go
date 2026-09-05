package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateStructure(t *testing.T) {
	t.Run("HappyPath", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp("", "project_test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir) // clean up

		projectName := "test_project"
		projectPath := filepath.Join(tempDir, projectName)

		err = CreateStructure(projectPath)
		if err != nil {
			t.Fatalf("CreateStructure failed: %v", err)
		}

		// Verify project directory exists
		if _, err := os.Stat(projectPath); os.IsNotExist(err) {
			t.Errorf("Project directory %s was not created", projectPath)
		}

		// Verify subdirectories
		expectedDirs := []string{"src", "tests", "docs", "config"}
		for _, dir := range expectedDirs {
			dirPath := filepath.Join(projectPath, dir)
			if info, err := os.Stat(dirPath); os.IsNotExist(err) || !info.IsDir() {
				t.Errorf("Expected directory %s was not created", dirPath)
			}
		}

		// Verify files and content
		readmePath := filepath.Join(projectPath, "README.md")
		readmeContent, err := os.ReadFile(readmePath)
		if err != nil {
			t.Errorf("Failed to read README.md: %v", err)
		} else {
			// Check that it contains the project path in README.md
			if !strings.Contains(string(readmeContent), projectPath) {
				t.Errorf("README.md does not contain expected project name/path")
			}
		}

		pkgPath := filepath.Join(projectPath, "package.json")
		pkgContent, err := os.ReadFile(pkgPath)
		if err != nil {
			t.Errorf("Failed to read package.json: %v", err)
		} else {
			if !strings.Contains(string(pkgContent), projectPath) {
				t.Errorf("package.json does not contain expected project name/path")
			}
		}
	})

	t.Run("ErrorPath_MkdirAll", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "project_test_err")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir) // clean up

		// Create a read-only directory
		readOnlyDir := filepath.Join(tempDir, "readonly")
		if err := os.Mkdir(readOnlyDir, 0444); err != nil {
			t.Fatalf("Failed to create readonly dir: %v", err)
		}

		// Attempt to create structure inside the read-only directory
		projectPath := filepath.Join(readOnlyDir, "test_project")
		err = CreateStructure(projectPath)
		if err == nil {
			t.Errorf("Expected an error when creating structure in a read-only directory, got nil")
		}
	})
}
