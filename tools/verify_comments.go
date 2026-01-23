package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/phravins/devcli/internal/boilerplate"
)

func main() {
	fmt.Println("Verifying Commented Snippets...")

	// Go up one level from scripts/ to root
	cwd, _ := os.Getwd()
	rootDir := filepath.Dir(cwd)
	if filepath.Base(cwd) != "scripts" {
		// If run from root, cwd is root
		rootDir = cwd
	}

	testDir := filepath.Join(rootDir, "test_snippets_comments")
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	mgr := boilerplate.NewManager(testDir)

	// Just spot check a few critical ones
	tests := []struct {
		name string
		lang string
		file string
	}{
		{"Auth System", "Go", "auth_handlers.go"},
		{"DB: PostgreSQL", "Go", "db_postgres.go"},
	}

	for _, tt := range tests {
		fmt.Printf("Test: %s... ", tt.name)
		_, err := mgr.GenerateSnippet(tt.name, tt.lang, testDir) // Ignored path return
		if err != nil {
			fmt.Printf("FAILED: %v\n", err)
			os.Exit(1)
		}

		// Verify file exists
		content, err := os.ReadFile(filepath.Join(testDir, tt.file))
		if err != nil {
			fmt.Printf("FAILED (Read Error): %v\n", err)
			os.Exit(1)
		}

		// Verify file has minimum size (implying comments are there)
		if len(content) < 100 {
			fmt.Printf("FAILED (Content too short)\n")
			os.Exit(1)
		}
		fmt.Println("PASSED")
	}

	fmt.Println("All Commented Snippets Verified!")
}
